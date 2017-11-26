package main

import (
	"bufio"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/shankj3/ocelot/util/ocelog"
	"github.com/shankj3/ocelot/util/ocenet"
	"github.com/shankj3/ocelot/util/storage"
	"io"
	"io/ioutil"
	"net/http"
	"time"
	"github.com/shankj3/ocelot/util/consulet"
	"fmt"
)

var (
	upgrader = websocket.Upgrader{}
	// todo: make sync.Once init for consulet, consul ocenet paths should all be in one file likely in consulet
	consulDonePath = "ci/builds/%s/done" //  %s is hash
	con = consulet.Default()
)
// modified from https://elithrar.github.io/article/custom-handlers-avoiding-globals/
type appContext struct {
	chanDict      *CD
	conf 		  *WerkerConf
	storage 	  storage.BuildOutputStorage
	readerCache   map[string] io.ReadCloser
}

type appHandler struct {
	*appContext
	H func(*appContext, http.ResponseWriter, *http.Request)
}

func (ah appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ah.H(ah.appContext, w, r)
}


func stream(a *appContext, w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	hash := vars["hash"]
	ocelog.Log().Debug(hash)
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		ocelog.IncludeErrField(err).Error("wtf?")
		return
	}
	defer ws.Close()
	bundleDone := make(chan int)
	infochan, ok := a.chanDict.CarefulValue(hash)
	if !ok {
		ocelog.Log().Debug("no info chan found")
		return
	}
	infoReader, infoWriter := getPipe(hash, a)
	// add BuildOutputStorage implementation, git hash,
	go writeBundle(infochan, infoWriter, infoReader, a.storage, hash)
	go pumpBundle(ws, infoReader, a.storage, hash, bundleDone)
	ocelog.Log().Debug("sending infoChan over web socket, waiting for the channel to be closed.")
	<-bundleDone
}

// getPipe looks up the hash in the appContext "readerCache" map to see if there is already a
// process dealing with this hash (i.e. multiple people go to build page). if there is,
// it returns the cached infoReader and a nil infoWriter which means that this request shouldn't result
// in writing to the Pipe.
// if not in map, create reader / writer from io.Pipe() and add to readerCache
func getPipe(hash string, a *appContext) (infoReader io.ReadCloser, infoWriter io.WriteCloser) {
	if m, ok := a.readerCache[hash]; !ok {
		infoReader, infoWriter = io.Pipe()
		a.readerCache[hash] = m
		return
	} else {
		infoReader = m
		return
	}
}

// pumpBundle writes build data to web socket
func pumpBundle(ws *websocket.Conn, infoReader io.Reader, persist storage.BuildOutputStorage, hash string, done chan int){
	var s *bufio.Scanner

	// determine whether to get from storage or off infoReader
	if CheckIfBuildDone(hash) {
		reader, err := persist.RetrieveReader(hash)
		if err != nil {
			ocelog.IncludeErrField(err).Error("could not retrieve persisted build data")
			return
		}
		s = bufio.NewScanner(reader)
	} else {
		ocelog.Log().Debug("pumping info reader data to web socket")
		s = bufio.NewScanner(infoReader)
	}
	// write to web socket
	for s.Scan() {
		ws.SetWriteDeadline(time.Now().Add(10*time.Second))
		if err := ws.WriteMessage(websocket.TextMessage, s.Bytes()); err != nil{
			ocelog.IncludeErrField(err).Error("could not write to web socket")
			ws.Close()
			break
		}
	}
	if s.Err() != nil {
		ocelog.IncludeErrField(s.Err()).Error("infoReader scan error")
	}
	ocelog.Log().Debug("finished pumping info reader data to web socket")
	defer func(){
		close(done)
		ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		ws.Close()
	}()
}

// writeBundle writes from infoChannel to writeCloser so multiple requests can access build data.
// when the infochan is closed, the reader is stored through the storage.BuildOutputStorage interface
// *this function will close both sides of the Pipe (w io.WriteCloser & r io.ReadCloser) after the infochan is closed.*
func writeBundle(infochan chan[]byte, w io.WriteCloser, r io.ReadCloser, persist storage.BuildOutputStorage, hash string){
	// if writer is not null, that means it was never put in the cache (see getPipe.)
	// prevents duping of readers on the same hash.
	if w != nil {
		for i := range infochan {
			newline := []byte("\n")
			w.Write(i)
			w.Write(newline)
		}
		w.Close()
		defer r.Close()
		bytez, err := ioutil.ReadAll(r)
		if err != nil {
			//todo: return error? set flag somewhere?
			ocelog.IncludeErrField(err).Error("could not read build data from info reader")
			return
		}
		if err = persist.Store(hash, bytez); err != nil {
			ocelog.IncludeErrField(err).Error("could not store build data to storage")
		} else {
			//todo: remove from cache

			// set to done
			if err := SetBuildDone(hash); err != nil {
				ocelog.IncludeErrField(err).Error("could not set build done")
			}
		}
	}
}

// TransportToCD reads off transport channel and adds the hash and infochan to the synced channel dict
func TransportToCD(tranChan chan *Transport, cd *CD){
	for i := range tranChan {
		ocelog.Log().Debugf("adding info channel for hash %s to map for streaming access.", i.Hash)
		if err := cd.CarefulPut(i.Hash, i.InfoChan); err != nil{
			ocelog.IncludeErrField(err).Error("could not add hash and info channel to map, " +
														"will not be able to stream results")
		}

	}
}

func SetBuildDone(gitHash string) error {
	// todo: add byte of 0/1.. have ot use binary library though and idk how to use that yet
	// and not motivated enough to do it right now
	err := con.AddKeyValue(fmt.Sprintf(consulDonePath, gitHash), []byte("true"))
	if err != nil {
		return err
	}
	return nil
}

func CheckIfBuildDone(gitHash string) bool {
	kv, err := con.GetKeyValue(fmt.Sprintf(consulDonePath, gitHash))
	if err != nil {
		// idk what we should be doing if the error is not nil, maybe panic? hope that never happens?
		return false
	}
	if kv != nil {
		return true
	}
	return false
}



func serveHome(w http.ResponseWriter, r *http.Request){
	http.ServeFile(w, r, "test.html")
}

func GetWerkConfig(conf *WerkerConf) *appContext{
	cd := NewCD()
	store := storage.NewFileBuildStorage("")
	readCache := make(map[string] io.ReadCloser)
	appctx := &appContext{chanDict: cd, conf: conf, storage: store, readerCache: readCache,}
	return appctx
}

func ServeMe(transportChan chan *Transport, conf *WerkerConf){
	appctx := GetWerkConfig(conf)
	go TransportToCD(transportChan, appctx.chanDict)
	muxi := mux.NewRouter()
	muxi.Handle("/ws/builds/{hash}", appHandler{appctx, stream}).Methods("GET")
	muxi.HandleFunc("/builds/{hash}", serveHome).Methods("GET")
	n := ocenet.InitNegroni("werker", muxi)
	n.Run(":"+conf.servicePort)
}
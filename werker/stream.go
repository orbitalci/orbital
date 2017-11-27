package main

import (
	"bufio"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/shankj3/ocelot/util/consulet"
	"github.com/shankj3/ocelot/util/ocelog"
	"github.com/shankj3/ocelot/util/ocenet"
	"github.com/shankj3/ocelot/util/storage"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	upgrader = websocket.Upgrader{}
	consulDonePath = "ci/builds/%s/done" //  %s is hash
)
// modified from https://elithrar.github.io/article/custom-handlers-avoiding-globals/
type appContext struct {
	conf 		  *WerkerConf
	storage 	  storage.BuildOutputStorage
	readerCache   *ReaderCache
	consul        *consulet.Consulet
}

// getPipe looks up the hash in the appContext "readerCache" map to see if there is already a
// process dealing with this hash (i.e. multiple people go to build page). if there is,
// it returns the cached infoReader and a nil infoWriter which means that this request shouldn't result
// in writing to the Pipe.
// if not in map, create reader / writer from io.Pipe() and add to readerCache
func (a *appContext) getPipe(hash string) (infoReader *io.PipeReader, infoWriter *io.PipeWriter) {
	if m, ok := a.readerCache.CarefulValue(hash); !ok {
		ocelog.Log().Debugf("could not find %s in cache", hash)
		var ir *io.PipeReader
		ir, infoWriter = io.Pipe()
		_ = a.readerCache.CarefulPut(hash, ir)
		infoReader = ir
		return
	} else {
		ocelog.Log().Debugf("found %s in reader cache", hash)
		// make a copy
		infoReader = &m
		return
	}
}

func (a *appContext)  SetBuildDone(gitHash string) error {
	// todo: add byte of 0/1.. have ot use binary library though and idk how to use that yet
	// and not motivated enough to do it right now
	err := a.consul.AddKeyValue(fmt.Sprintf(consulDonePath, gitHash), []byte("true"))
	if err != nil {
		return err
	}
	return nil
}

func (a *appContext) CheckIfBuildDone(gitHash string) bool {
	kv, err := a.consul.GetKeyValue(fmt.Sprintf(consulDonePath, gitHash))
	if err != nil {
		// idk what we should be doing if the error is not nil, maybe panic? hope that never happens?
		return false
	}
	if kv != nil {
		return true
	}
	return false
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
	infoReader, ok := a.readerCache.CarefulValue(hash)
	if !ok {
		// todo, write "try later" to socket
		ocelog.Log().Error("try again later, could not find any data")
		return
	}
	if &infoReader == nil {
		// this shouldn't happen....
		ocelog.Log().Fatal("wtf?")
	}
	go pumpBundle(ws, &infoReader, a, hash, bundleDone)
	ocelog.Log().Debug("sending infoChan over web socket, waiting for the channel to be closed.")
	<-bundleDone
}

// pumpBundle writes build data to web socket
func pumpBundle(ws *websocket.Conn, infoReader *io.PipeReader, appCtx *appContext, hash string, done chan int){
	var s *bufio.Scanner
	var copied = *infoReader
	// determine whether to get from storage or off infoReader
	if appCtx.CheckIfBuildDone(hash) {
		ocelog.Log().Debugf("build %s is done, getting from appCtx", hash)
		reader, err := appCtx.storage.RetrieveReader(hash)
		if err != nil {
			ocelog.IncludeErrField(err).Error("could not retrieve persisted build data")
			return
		}
		s = bufio.NewScanner(reader)
	} else {
		ocelog.Log().Debug("pumping info reader data to web socket")
		s = bufio.NewScanner(&copied)
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

// writeInfoChanToCache is what processes transport objects that come from the transport channel (objects get created
// 	and sent when a new build is pulled off the queue). a pipe is created, and the readerCache is populated for that
// 	git hash and the io.PipeReader
// 	the info chan is written to the io.PipeWriter, then stored using the storage config in appCtx for persistence.
// 	consul is updated with build done status, and readerCache entry is removed
func writeInfoChanToCache(transport  *Transport, appCtx *appContext){
	// create PipeReader and PipeWriter objects for handling InfoChan data
	r, w := io.Pipe()
	// add to readerCache
	appCtx.readerCache.CarefulPut(transport.Hash, r)
	ocelog.Log().Debugf("writing infochan data for %s", transport.Hash)
	for i := range transport.InfoChan {
		newline := []byte("\n")
		w.Write(i)
		w.Write(newline)
	}
	w.Close()
	defer r.Close()
	// persist everything
	ocelog.Log().Debug("storing in <>") // todo: add String() method to storage interface
	bytez, err := ioutil.ReadAll(r)
	if err != nil {
		//todo: return error? set flag somewhere?
		ocelog.IncludeErrField(err).Fatal("could not read build data from info reader")
		return
	}
	err = appCtx.storage.Store(transport.Hash, bytez)
	if err != nil {
		ocelog.IncludeErrField(err).Fatal("could not store build data to storage")
		return
	}
	// get rid of hash from cache, set build done in consul
	if err := appCtx.SetBuildDone(transport.Hash); err != nil {
		ocelog.IncludeErrField(err).Error("could not set build done")
	}
	ocelog.Log().Debugf("removing hash %s from readerCache and channelDict", transport.Hash)
	appCtx.readerCache.CarefulRm(transport.Hash)

}

func CacheProcessor(transpo chan *Transport, appCtx *appContext){
	for i := range transpo {
		ocelog.Log().Debugf("adding info channel for hash %s to map for streaming access.", i.Hash)
		writeInfoChanToCache(i, appCtx)
	}
}


func serveHome(w http.ResponseWriter, r *http.Request){
	http.ServeFile(w, r, "test.html")
}

func GetWerkConfig(conf *WerkerConf) *appContext{
	store := storage.NewFileBuildStorage("")
	appCtx := &appContext{ conf: conf, storage: store, readerCache: NewReaderCache(), consul: consulet.Default()}
	return appCtx
}

func ServeMe(transportChan chan *Transport, conf *WerkerConf){
	ocelog.Log().Debug("started serving routine for streaming data")
	appctx := GetWerkConfig(conf)
	go CacheProcessor(transportChan, appctx)
	muxi := mux.NewRouter()
	muxi.Handle("/ws/builds/{hash}", appHandler{appctx, stream}).Methods("GET")
	muxi.HandleFunc("/builds/{hash}", serveHome).Methods("GET")
	n := ocenet.InitNegroni("werker", muxi)
	n.Run(":"+conf.servicePort)
	ocelog.Log().Info("QUITTING")
}
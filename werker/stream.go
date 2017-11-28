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
	buildInfo     map[string]*buildDatum
	consul        *consulet.Consulet
}


type buildDatum struct {
	buildData [][]byte
	done      chan int
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


func stream(ctx interface{}, w http.ResponseWriter, r *http.Request){
	a := ctx.(*appContext)
	vars := mux.Vars(r)
	hash := vars["hash"]
	ocelog.Log().Debug(hash)
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		ocelog.IncludeErrField(err).Error("wtf?")
		return
	}
	defer ws.Close()
	pumpDone := make(chan int)

	go pumpBundle(ws, a, hash, pumpDone)
	ocelog.Log().Debug("sending infoChan over web socket, waiting for the channel to be closed.")
	<-pumpDone
}

// pumpBundle writes build data to web socket
func pumpBundle(ws *websocket.Conn, appCtx *appContext, hash string, done chan int){
	// determine whether to get from storage or off infoReader
	if appCtx.CheckIfBuildDone(hash) {
		ocelog.Log().Debugf("build %s is done, getting from appCtx", hash)
		reader, err := appCtx.storage.RetrieveReader(hash)
		if err != nil {
			ocelog.IncludeErrField(err).Error("could not retrieve persisted build data")
			return
		}
		s := bufio.NewScanner(reader)
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
	} else {
		ocelog.Log().Debug("pumping info reader data to web socket")
		buildInfo := appCtx.buildInfo[hash]
		err := streamFromArray(buildInfo, ws)
		if err != nil {
			ocelog.IncludeErrField(err).Error("could not stream from array!")
		}
	}
	ocelog.Log().Debug("finished pumping info reader data to web socket")
	defer func(){
		close(done)
		ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		ws.Close()
	}()
}

func streamFromArray(buildInfo *buildDatum, ws *websocket.Conn) (err error){
	var index int
	for {
		buildData := buildInfo.buildData[index:]
		select {
		case <-buildInfo.done:
			return nil
		default:
			index, err = iterateOverBuildData(buildData, ws)
			if err != nil {
				return err
			}
		}
		ocelog.Log().Debugf("idk what a good description would be for what is happening....")
		time.Sleep(1 * time.Second)
	}

}


func iterateOverBuildData(data [][]byte, ws *websocket.Conn) (int, error) {
	var index int
	for index, dataLine := range data {
		ws.SetWriteDeadline(time.Now().Add(10*time.Second))
		if err := ws.WriteMessage(websocket.TextMessage, dataLine); err != nil {
			ocelog.IncludeErrField(err).Error("could not write to web socket")
			ws.Close()
			return index, err
		}
	}
	return index, nil
}

// todo: update docstring
// writeInfoChanToCache is what processes transport objects that come from the transport channel (objects get created
// 	and sent when a new build is pulled off the queue). a pipe is created, and the readerCache is populated for that
// 	git hash and the io.PipeReader
// 	the info chan is written to the io.PipeWriter, then stored using the storage config in appCtx for persistence.
// 	consul is updated with build done status, and readerCache entry is removed
func writeInfoChanToCache(transport  *Transport, appCtx *appContext){
	r, w := io.Pipe()
	defer r.Close()

	var dataSlice [][]byte
	build := &buildDatum{dataSlice, make(chan int),}
	appCtx.buildInfo[transport.Hash] = build
	ocelog.Log().Debugf("writing infochan data for %s", transport.Hash)
	// todo: change the worker to act as grpc server
	// to expose method that lets you get stream as commit
	// keep in memory of output of build
	// create a new stream @ every request
	for i := range transport.InfoChan {
		// for streaming
		ocelog.Log().Debug("ayy")
		build.buildData = append(build.buildData, i)
		// for storing
		newline := []byte("\n")
		w.Write(i)
		w.Write(newline)
	}
	w.Close()
	ocelog.Log().Debug("supposedly done???")
	build.done <- 0
	bytez, err := ioutil.ReadAll(r)
	if err != nil {
		ocelog.IncludeErrField(err).Error("could not read off PipeReader")
		return
	}
	err = appCtx.storage.Store(transport.Hash, bytez)
	if err != nil {
		ocelog.IncludeErrField(err).Error("could not store build data to storage")
		return
	}
	// get rid of hash from cache, set build done in consul
	if err := appCtx.SetBuildDone(transport.Hash); err != nil {
		ocelog.IncludeErrField(err).Error("could not set build done")
	}
	ocelog.Log().Debugf("removing hash %s from readerCache and channelDict", transport.Hash)
	delete(appCtx.buildInfo, transport.Hash)

}

func cacheProcessor(transpo chan *Transport, appCtx *appContext){
	for i := range transpo {
		ocelog.Log().Debugf("adding info channel for hash %s to map for streaming access.", i.Hash)
		go writeInfoChanToCache(i, appCtx)
	}
}


func serveHome(w http.ResponseWriter, r *http.Request){
	http.ServeFile(w, r, "test.html")
}

func GetWerkConfig(conf *WerkerConf) *appContext{
	store := storage.NewFileBuildStorage("")
	appCtx := &appContext{ conf: conf, storage: store, buildInfo: make(map[string]*buildDatum), consul: consulet.Default()}
	return appCtx
}

func ServeMe(transportChan chan *Transport, conf *WerkerConf){
	ocelog.Log().Debug("started serving routine for streaming data")
	appctx := GetWerkConfig(conf)
	go cacheProcessor(transportChan, appctx)
	muxi := mux.NewRouter()
	muxi.Handle("/ws/builds/{hash}", &ocenet.AppContextHandler{appctx, stream}).Methods("GET")
	muxi.HandleFunc("/builds/{hash}", serveHome).Methods("GET")
	n := ocenet.InitNegroni("werker", muxi)
	n.Run(":"+conf.servicePort)
	ocelog.Log().Info("QUITTING")
}
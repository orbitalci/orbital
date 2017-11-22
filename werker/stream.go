package main

import (
	"bufio"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/shankj3/ocelot/util/ocelog"
	"github.com/shankj3/ocelot/util/ocenet"
	"io"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{}

// modified from https://elithrar.github.io/article/custom-handlers-avoiding-globals/
type appContext struct {
	chanDict      *CD
	conf 		  *WerkerConf
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
	ocelog.Log().Debug(vars["hash"])
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		ocelog.IncludeErrField(err).Error("wtf?")
		return
	}
	defer ws.Close()
	bundleDone := make(chan int)
	infochan, ok := a.chanDict.CarefulValue(vars["hash"])
	if !ok {
		ocelog.Log().Debug("no info chan found")
		return
	}
	infoReader, infoWriter := io.Pipe()
	defer infoReader.Close()
	defer infoWriter.Close()

	go writeBundle(infochan, infoWriter)
	go pumpBundle(ws, infoReader, bundleDone)
	ocelog.Log().Debug("sending infoChan over websocket, waiting for the channel to be closed.")
	<-bundleDone
}

// todo: need another goroutine to read off the infochan so that more than one
// page can hit the site
func pumpBundle(ws *websocket.Conn, infoReader io.Reader, done chan int){
	defer func(){}()
	s := bufio.NewScanner(infoReader)
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
	close(done)
	ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	ws.Close()
}


func writeBundle(infochan chan[]byte, w io.WriteCloser){
	for i := range infochan {
		newline := []byte("\n")
		w.Write(i)
		w.Write(newline)
	}
	w.Close()
}


func TransportToCD(tranChan chan *Transport, cd *CD){
	for i := range tranChan {
		ocelog.Log().Debugf("adding info channel for hash %s to map for streaming access.", i.Hash)
		if err := cd.CarefulPut(i.Hash, i.InfoChan); err != nil{
			ocelog.IncludeErrField(err).Error("could not add hash and info channel to map, " +
														"will not be able to stream results")
		}

	}
}

func serveHome(w http.ResponseWriter, r *http.Request){
	http.ServeFile(w, r, "test.html")
}

func ServeMe(transportChan chan *Transport, conf *WerkerConf){
	cd := NewCD()
	appctx := &appContext{chanDict: cd, conf: conf,}
	go TransportToCD(transportChan, cd)
	muxi := mux.NewRouter()
	muxi.Handle("/ws/builds/{hash}", appHandler{appctx, stream}).Methods("GET")
	muxi.HandleFunc("/builds/{hash}", serveHome).Methods("GET")
	n := ocenet.InitNegroni("werker", muxi)
	n.Run(":"+conf.servicePort)
}
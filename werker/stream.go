package main

import (
	"github.com/gorilla/websocket"
	"github.com/shankj3/ocelot/util/ocelog"
	"net/http"
)

var upgrader = websocket.Upgrader{}

func stream(w http.ResponseWriter, r *http.Request){
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		ocelog.IncludeErrField(err).Error("wtf?")
		return
	}
	defer ws.Close()
	bundleDone := make(chan int)
	pumpBundle(ws, bundleDone)
}

func pumpBundle(ws *websocket.Conn, done chan int){

}

func ServeMe(){
	conf, err := GetConf()
	if err != nil {
		ocelog.Log().Fatal("cannot get configuration")
	}

}
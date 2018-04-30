package werker

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	ocelog "github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"
)

var (
	upgrader = websocket.Upgrader{}
)

func addHandlers(muxi *mux.Router, werkData *WerkerContext) {
	muxi.Handle("/ws/builds/{hash}", &ocenet.AppContextHandler{werkData, stream}).Methods("GET")
	muxi.HandleFunc("/builds/{hash}", serveHome).Methods("GET")
	//muxi.HandleFunc("/DUMP", werkStream.dumpData).Methods("GET")

	//if we're in dev mode, serve everything out of test-fixtures at /dev
	mode := os.Getenv("ENV")
	if strings.EqualFold(mode, "dev") {
		muxi.PathPrefix("/dev/").Handler(http.StripPrefix("/dev/", http.FileServer(http.Dir("./dev"))))
	}

	//serve up zip files that spawned containers need
	muxi.HandleFunc("/do_things.tar", func(w http.ResponseWriter, r *http.Request) {
		ocelog.Log().Debug("serving up zip files from s3")
		http.Redirect(w, r, "https://s3-us-west-2.amazonaws.com/ocelotty/werker_files.tar", 301)
	})

	muxi.HandleFunc("/kubectl", func(w http.ResponseWriter, r *http.Request) {
		ocelog.Log().Debug("serving up kubectl binary from googleapis")
		http.Redirect(w, r, "https://storage.googleapis.com/kubernetes-release/release/v1.9.6/bin/linux/amd64/kubectl", 301)
	})

}

func serveHome(w http.ResponseWriter, r *http.Request) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("no caller???? ")
	}
	fmt.Println("FILENAME ", path.Dir(filename)+"/test.html")
	http.ServeFile(w, r, path.Dir(filename)+"/test.html")
}

func stream(ctx interface{}, w http.ResponseWriter, r *http.Request) {
	a := ctx.(*WerkerContext)
	vars := mux.Vars(r)
	hash := vars["hash"]
	ocelog.Log().Debug(hash)
	ws, err := ocenet.Upgrade(upgrader, w, r)
	if err != nil {
		ocelog.IncludeErrField(err).Error("wtf?")
		return
	}
	//defer ws.Close()
	pumpDone := make(chan int)

	go a.streamPack.PumpBundle(ws, hash, pumpDone)
	ocelog.Log().Debug("sending infoChan over web socket, waiting for the channel to be closed.")
	<-pumpDone
}

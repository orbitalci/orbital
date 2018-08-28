package werker

import (
	"fmt"
	"net/http"
	"path"
	"runtime"





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

	//todo: don't think we are going this way anymore... delete?
	//if we're in dev mode, serve everything out of test-fixtures at /dev
	//if werkData.Dev {
	//	muxi.PathPrefix("/dev/").Handler(http.StripPrefix("/dev/", http.FileServer(http.Dir("./dev"))))
	//}
	//serve up zip files that spawned containers need
	muxi.HandleFunc("/do_things.tar", func(w http.ResponseWriter, r *http.Request) {
		if werkData.Dev {
			ocelog.Log().Info("DEV MODE, SERVING WERKER FILES LOCALLY")
			_, filename, _, ok := runtime.Caller(0)
			if !ok {
				panic("no caller???? ")
			}
			http.ServeFile(w, r, path.Dir(filename)+"/werker_files.tar")
		} else {
			// todo: move to the base64 encode then echo to file method, this is frustrating
			ocelog.Log().Debug("serving up zip files from s3")
			/// todo: change this back!!!!
			http.Redirect(w, r, "https://s3-us-west-2.amazonaws.com/ocelotty/werker_files_dev.tar", 301)
		}
	})

	// todo: THESE ARE ALL LINUX-SPECIFIC!
	muxi.HandleFunc("/kubectl", func(w http.ResponseWriter, r *http.Request) {
		ocelog.Log().Debug("serving up kubectl binary from googleapis")
		http.Redirect(w, r, "https://storage.googleapis.com/kubernetes-release/release/v1.9.6/bin/linux/amd64/kubectl", 301)
	})
	muxi.HandleFunc("/helm.tar.gz", func(w http.ResponseWriter, r *http.Request) {
		ocelog.Log().Debug("serving up helm binary from googleapis")
		http.Redirect(w, r, "https://storage.googleapis.com/kubernetes-helm/helm-v2.10.0-linux-amd64.tar.gz", 301)
	})
	muxi.HandleFunc("/mc", func(w http.ResponseWriter, r *http.Request) {
		ocelog.Log().Debug("serving up mc binary")
		http.Redirect(w, r, "https://dl.minio.io/client/mc/release/linux-amd64/mc", 301)
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

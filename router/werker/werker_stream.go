package werker

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	ocelog "github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"
	"github.com/shankj3/ocelot/build/streamer"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"
	"google.golang.org/grpc"
)

var (
	upgrader = websocket.Upgrader{}
)

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

//ServeMe will start HTTP Server as needed for streaming build output by hash
func ServeMe(transportChan chan *models.Transport, buildCtxChan chan *models.BuildContext, conf *models.WerkerFacts, store storage.OcelotStorage) {
	// todo: defer a recovery here
	werkStream := getWerkerContext(conf, store)
	streamPack := streamer.GetStreamPack(werkStream.BuildContexts, werkStream.store, werkStream.consul)
	werkStream.streamPack = streamPack
	ocelog.Log().Debug("saving build info channels to in memory map")
	go streamPack.ListenTransport(transportChan)
	go streamPack.ListenBuilds(buildCtxChan, sync.Mutex{})

	ocelog.Log().Info("serving websocket on port: ", conf.ServicePort)
	muxi := mux.NewRouter()
	muxi.Handle("/ws/builds/{hash}", &ocenet.AppContextHandler{werkStream, stream}).Methods("GET")
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

	n := ocenet.InitNegroni("werker", muxi)
	go n.Run(":" + conf.ServicePort)

	//start grpc server
	ocelog.Log().Info("serving grpc streams of build data on port: ", conf.GrpcPort)
	con, err := net.Listen("tcp", ":"+conf.GrpcPort)
	if err != nil {
		ocelog.Log().Fatal("womp womp")
	}

	grpcServer := grpc.NewServer()
	werkerServer := NewWerkerServer(werkStream)
	pb.RegisterBuildServer(grpcServer, werkerServer)
	go grpcServer.Serve(con)
	go func() {
		ocelog.Log().Info(http.ListenAndServe(":6060", nil))
	}()
}

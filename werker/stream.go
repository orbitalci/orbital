package werker

import (
	consulet "bitbucket.org/level11consulting/go-til/consul"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"bitbucket.org/level11consulting/ocelot/util/streamer"
	"bitbucket.org/level11consulting/ocelot/werker/protobuf"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
)

var (
	upgrader       = websocket.Upgrader{}
	consulDonePath = "ci/builds/%s/done" //  %s is hash
)

// modified from https://elithrar.github.io/article/custom-handlers-avoiding-globals/
type werkerStreamer struct {
	conf      *WerkerConf
	storage   storage.BuildOutputStorage
	buildInfo map[string]*buildDatum
	consul    *consulet.Consulet
}

type buildDatum struct {
	buildData [][]byte
	done      bool
}

func (b *buildDatum) GetData() [][]byte{
	return b.buildData
}

func (b *buildDatum) CheckDone() bool {
	return b.done
}


func (w *werkerStreamer) SetBuildDone(gitHash string) error {
	// todo: add byte of 0/1.. have ot use binary library though and idk how to use that yet
	// and not motivated enough to do it right now
	err := w.consul.AddKeyValue(fmt.Sprintf(consulDonePath, gitHash), []byte("true"))
	if err != nil {
		return err
	}
	return nil
}

func (w *werkerStreamer) CheckIfBuildDone(gitHash string) bool {
	kv, err := w.consul.GetKeyValue(fmt.Sprintf(consulDonePath, gitHash))
	if err != nil {
		// idk what we should be doing if the error is not nil, maybe panic? hope that never happens?
		return false
	}
	if kv != nil {
		return true
	}
	return false
}

func (w *werkerStreamer) BuildInfo(request *protobuf.Request, stream protobuf.Build_BuildInfoServer) error {
	resp := &protobuf.Response{
		OutputLine: request.Hash,
	}
	stream.Send(resp)
	stream.Send(&protobuf.Response{
		OutputLine: w.conf.WerkerName,
	})
	pumpDone := make(chan int)
	streamable := &protobuf.BuildStreamableServer{Server: stream}
	go pumpBundle(streamable, w, request.Hash, pumpDone)
	<-pumpDone
	return nil
}


func stream(ctx interface{}, w http.ResponseWriter, r *http.Request) {
	a := ctx.(*werkerStreamer)
	vars := mux.Vars(r)
	hash := vars["hash"]
	ocelog.Log().Debug(hash)
	//ws, err := upgrader.Upgrade(w, r, nil)
	ws, err := ocenet.Upgrade(upgrader, w, r)
	if err != nil {
		ocelog.IncludeErrField(err).Error("wtf?")
		return
	}
	//defer ws.Close()
	pumpDone := make(chan int)

	go pumpBundle(ws, a, hash, pumpDone)
	ocelog.Log().Debug("sending infoChan over web socket, waiting for the channel to be closed.")
	<-pumpDone
}


// pumpBundle writes build data to web socket
func pumpBundle(stream streamer.Streamable, appCtx *werkerStreamer, hash string, done chan int) {
	// determine whether to get from storage or off infoReader
	//if appCtx.CheckIfBuildDone(hash) {
	//	ocelog.Log().Debugf("build %s is done, getting from appCtx", hash)
	// // err := streamFromStorage(appCtx, hash, streamer)
	//  err := streamer.StreamFromStorage(appCtx.storage, stream, hash)
	//	if err != nil {
	//		ocelog.IncludeErrField(err).Error("error retrieving from storage")
	//	}
	//} else {
	//	ocelog.Log().Debug("pumping info array data to web socket")
		buildInfo, ok := appCtx.buildInfo[hash]
		if ok {
			err := streamer.StreamFromArray(buildInfo, stream, ocelog.Log().Debug)
			if err != nil {
				ocelog.IncludeErrField(err).Error("could not stream from array!")
			}
			ocelog.Log().Debug("streamed build data from array")
		} else {
			stream.SendError([]byte("did not find hash in current streaming data and the build was not marked as done"))
		}
	//}
	//defer streamer.Finish()
}

// writeInfoChanToInMemMap is what processes transport objects that come from the transport channel (objects get created
// 	and sent when a new build is pulled off the queue).
// 	the info channel is written to an array which is put in a map in the appCtx along with a done channel so
//  there is a way to see when the array will not be written to anymore
//  when the info channel is closed and the loop finishes, all the data is written to the storage defined in the
//  appCtx, the done flag is written to consul, and the array is removed from the map
func writeInfoChanToInMemMap(transport *Transport, appCtx *werkerStreamer) {
	// question: does this support unicode?
	var dataSlice [][]byte
	build := &buildDatum{dataSlice, false}
	appCtx.buildInfo[transport.Hash] = build
	ocelog.Log().Debugf("writing infochan data for %s", transport.Hash)
	for i := range transport.InfoChan {
		build.buildData = append(build.buildData, i)
	}
	ocelog.Log().Debug("done with build ", transport.Hash)
	err := appCtx.storage.StoreLines(transport.Hash, build.buildData)
	// even if it didn't store properly, we need to set the build in the map as "done" so
	// that the streams that connect when the build is still happening know to close the connection
	build.done = true
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

func cacheProcessor(transpo chan *Transport, appCtx *werkerStreamer) {
	for i := range transpo {
		ocelog.Log().Debugf("adding info channel for hash %s to map for streaming access.", i.Hash)
		go writeInfoChanToInMemMap(i, appCtx)
	}
}

//****WARNING**** this assumes you're inside of /cmd/werker directory
func serveHome(w http.ResponseWriter, r *http.Request) {
	//pwd, _ := os.Getwd()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("no caller???? ")
	}
	fmt.Println("FILENAME ",  path.Dir(filename) + "/test.html")
	http.ServeFile(w, r, path.Dir(filename) + "/test.html")
}

func getWerkerStreamer(conf *WerkerConf) *werkerStreamer {
	store := storage.NewFileBuildStorage("")
	werkerConsul, err := consulet.Default()
	if err != nil {
		ocelog.IncludeErrField(err)
	}
	werkerStreamer := &werkerStreamer{
		conf:      conf,
		storage:   store,
		buildInfo: make(map[string]*buildDatum),
		consul:    werkerConsul}
	return werkerStreamer
}

//ServeMe will start HTTP Server as needed for streaming build output by hash
func ServeMe(transportChan chan *Transport, conf *WerkerConf) {
	werkStream := getWerkerStreamer(conf)
	ocelog.Log().Debug("saving build info channels to in memory map")
	go cacheProcessor(transportChan, werkStream)

	ocelog.Log().Info("serving websocket on port: ", conf.servicePort)
	muxi := mux.NewRouter()
	muxi.Handle("/ws/builds/{hash}", &ocenet.AppContextHandler{werkStream, stream}).Methods("GET")
	muxi.HandleFunc("/builds/{hash}", serveHome).Methods("GET")

	//if we're in dev mode, serve everything out of test-fixtures at /dev
	mode := os.Getenv("ENV")
	if strings.EqualFold(mode, "dev") {
		muxi.PathPrefix("/dev/").Handler(http.StripPrefix("/dev/", http.FileServer(http.Dir("./dev"))))
	}

	n := ocenet.InitNegroni("werker", muxi)
	go n.Run(":" + conf.servicePort)

	ocelog.Log().Info("serving grpc streams of build data on port: ", conf.grpcPort)
	con, err := net.Listen("tcp", ":"+conf.grpcPort)
	if err != nil {
		ocelog.Log().Fatal("womp womp")
	}
	grpcServer := grpc.NewServer()
	protobuf.RegisterBuildServer(grpcServer, werkStream)
	go grpcServer.Serve(con)

}

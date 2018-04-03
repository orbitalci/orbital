package werker

import (
	_ "net/http/pprof"
	consulet "bitbucket.org/level11consulting/go-til/consul"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	rt "bitbucket.org/level11consulting/ocelot/util/buildruntime"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"bitbucket.org/level11consulting/ocelot/util/storage/models"
	"bitbucket.org/level11consulting/ocelot/util/streamer"
	"bitbucket.org/level11consulting/ocelot/werker/protobuf"
	"bytes"
	"encoding/json"
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
	"sync"
	"time"
	"bitbucket.org/level11consulting/ocelot/werker/config"
)

var (
	upgrader = websocket.Upgrader{}
)

type WerkerContext struct {
	BuildContexts map[string]*BuildContext
	Conf          *config.WerkerConf

	out       storage.BuildOut
	sum       storage.BuildSum
	buildInfo map[string]*buildDatum
	consul    *consulet.Consulet
}

func (w *WerkerContext) dumpData(wr http.ResponseWriter, r *http.Request) {
	ocelog.Log().Info("writing out data for buildInfo")
	wr.Header().Set("content-type", "application/json")
	dataMap := make(map[string]int)
	dataMap["time"] = int(time.Now().Unix())
	wr.WriteHeader(http.StatusOK)
	for hash, bytearray := range w.buildInfo {
		dataMap[hash] = len(bytearray.GetData())
	}
	bit, err := json.Marshal(dataMap)
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't marshal for dump")
		wr.WriteHeader(http.StatusInternalServerError)
		return
	}
	wr.Write(bit)
}

func getWerkerContext(conf *config.WerkerConf, store storage.OcelotStorage) *WerkerContext {
	werkerConsul, err := consulet.Default()
	if err != nil {
		ocelog.IncludeErrField(err)
	}
	werkerCtx := &WerkerContext{
		BuildContexts: make(map[string]*BuildContext),
		Conf:          conf,

		out:       store,
		sum:       store,
		buildInfo: make(map[string]*buildDatum),
		consul:    werkerConsul}
	return werkerCtx
}

type buildDatum struct {
	sync.Mutex
	buildData [][]byte
	done      bool
}

func (b *buildDatum) GetData() [][]byte {
	// just an idea for if we get _more_ problems
	//var copied [][]byte
	//b.Lock()
	//defer b.Unlock()
	//copyLen := copy(b.buildData, copied)
	//if copyLen != len(b.buildData) {
	//	fmt.Println("LENGTHS NOT THE SAME! ", copyLen, leN(b.buildData))
	//}
	return b.buildData
}

func (b *buildDatum) CheckDone() bool {
	return b.done
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

	go pumpBundle(ws, a, hash, pumpDone)
	ocelog.Log().Debug("sending infoChan over web socket, waiting for the channel to be closed.")
	<-pumpDone
}

// pumpBundle writes build data to web socket
func pumpBundle(stream streamer.Streamable, appCtx *WerkerContext, hash string, done chan int) {
	defer func() {
		if r := recover(); r != nil {
			ocelog.Log().WithField("recover", r).Error("recovered from a panic in pumpBundle!!")
		}
	}()
	// determine whether to get from out or off infoReader
	if rt.CheckIfBuildDone(appCtx.consul, appCtx.sum, hash) {
		ocelog.Log().Debugf("build %s is done, getting from appCtx", hash)
		latestSummary, err := appCtx.sum.RetrieveLatestSum(hash)
		if err != nil {
			ocelog.IncludeErrField(err).Error("could not get latest build from storage")
		} else {
			err = streamer.StreamFromStorage(appCtx.out, stream, latestSummary.BuildId)
			if err != nil {
				ocelog.IncludeErrField(err).Error("error retrieving from storage")
			}
		}
	} else {
		ocelog.Log().Debug("pumping info array data to web socket")
		buildInfo, ok := appCtx.buildInfo[hash]
		ocelog.Log().Debug("length of array to stream is %d", len(buildInfo.GetData()))
		if ok {
			err := streamer.StreamFromArray(buildInfo, stream, ocelog.Log())
			if err != nil {
				ocelog.IncludeErrField(err).Error("could not stream from array!")
				return
			}
			ocelog.Log().Debug("streamed build data from array")
		} else {
			stream.SendError([]byte("did not find hash in current streaming data and the build was not marked as done"))
		}
	}
	defer stream.Finish(done)
}

// processTransport deals with adding info to consul, and calling writeInfoChanToInMemMap
func processTransport(transport *Transport, appCtx *WerkerContext) {
	writeInfoChanToInMemMap(transport, appCtx)
	// get rid of hash from cache, set build done in consul
	//if err := rt.SetBuildDone(appCtx.consul, transport.Hash); err != nil {
	//	ocelog.IncludeErrField(err).Error("could not set build done")
	//}
	ocelog.Log().Debugf("removing hash %s from readerCache, channelDict, and consul", transport.Hash)
	delete(appCtx.buildInfo, transport.Hash)
	if err := rt.Delete(appCtx.consul, transport.Hash); err != nil {
		ocelog.IncludeErrField(err).Error("could not recursively delete values from consul")
	}

}

// writeInfoChanToInMemMap is what processes transport objects that come from the transport channel (objects get created
// 	and sent when a new build is pulled off the queue).
// 	the info channel is written to an array which is put in a map in the appCtx along with a done channel so
//  there is a way to see when the array will not be written to anymore
//  when the info channel is closed and the loop finishes, all the data is written to the out defined in the
//  appCtx, the done flag is written to consul, and the array is removed from the map
func writeInfoChanToInMemMap(transport *Transport, appCtx *WerkerContext) {
	var dataSlice [][]byte
	build := &buildDatum{buildData: dataSlice, done: false}
	appCtx.buildInfo[transport.Hash] = build
	ocelog.Log().Debugf("writing infochan data for %s", transport.Hash)
	for i := range transport.InfoChan {
		build.Lock()
		build.buildData = append(build.buildData, i)
		build.Unlock()
		// i think wihtout this it eats all the cpu..
		time.Sleep(time.Millisecond)
	}
	ocelog.Log().Debug("done with build ", transport.Hash)
	out := &models.BuildOutput{
		BuildId: transport.DbId,
		Output:  bytes.Join(build.buildData, []byte("\n")),
	}
	//ocelog.Log().Debug(string(len(out.Output)))
	err := appCtx.out.AddOut(out)
	// even if it didn't store properly, we need to set the build in the map as "done" so
	// that the streams that connect when the build is still happening know to close the connection
	build.done = true
	if err != nil {
		ocelog.IncludeErrField(err).Error("could not store build data to storage")
	}
}

func listenTransport(transpo chan *Transport, appCtx *WerkerContext) {
	for i := range transpo {
		ocelog.Log().Debugf("adding info channel for hash %s to map for streaming access.", i.Hash)
		go processTransport(i, appCtx)
	}
}

func listenBuilds(buildsChan chan *BuildContext, appCtx *WerkerContext, mapLock sync.Mutex) {
	for newBuild := range buildsChan {
		mapLock.Lock()
		ocelog.Log().Debugf("got new build context for %s", newBuild.Hash)
		appCtx.BuildContexts[newBuild.Hash] = newBuild
		mapLock.Unlock()
		go contextCleanup(newBuild, appCtx, mapLock)
	}
}

func contextCleanup(buildCtx *BuildContext, appCtx *WerkerContext, mapLock sync.Mutex) {
	select {
		case <-buildCtx.Context.Done():
			mapLock.Lock()
			ocelog.Log().Debugf("build for hash %s is complete", buildCtx.Hash)
			if _, ok := appCtx.BuildContexts[buildCtx.Hash]; ok {
				delete(appCtx.BuildContexts, buildCtx.Hash)
			}
			mapLock.Lock()
			return
	}
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("no caller???? ")
	}
	fmt.Println("FILENAME ", path.Dir(filename)+"/test.html")
	http.ServeFile(w, r, path.Dir(filename)+"/test.html")
}

//ServeMe will start HTTP Server as needed for streaming build output by hash
func ServeMe(transportChan chan *Transport, buildCtxChan chan *BuildContext, conf *config.WerkerConf, store storage.OcelotStorage) {
	// todo: defer a recovery here
	werkStream := getWerkerContext(conf, store)
	ocelog.Log().Debug("saving build info channels to in memory map")
	go listenTransport(transportChan, werkStream)
	go listenBuilds(buildCtxChan, werkStream, sync.Mutex{})

	ocelog.Log().Info("serving websocket on port: ", conf.ServicePort)
	muxi := mux.NewRouter()
	muxi.Handle("/ws/builds/{hash}", &ocenet.AppContextHandler{werkStream, stream}).Methods("GET")
	muxi.HandleFunc("/builds/{hash}", serveHome).Methods("GET")
	muxi.HandleFunc("/DUMP", werkStream.dumpData).Methods("GET")

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
	protobuf.RegisterBuildServer(grpcServer, werkerServer)
	go grpcServer.Serve(con)
	go func() {
		ocelog.Log().Info(http.ListenAndServe(":6060", nil))
	}()

}

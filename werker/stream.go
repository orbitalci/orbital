package werker

import (
	consulet "bitbucket.org/level11consulting/go-til/consul"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	rt "bitbucket.org/level11consulting/ocelot/util/buildruntime"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"bitbucket.org/level11consulting/ocelot/util/storage/models"
	"bitbucket.org/level11consulting/ocelot/util/streamer"
	"bitbucket.org/level11consulting/ocelot/werker/protobuf"
	"bytes"
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
)

var (
	upgrader = websocket.Upgrader{}
)

// modified from https://elithrar.github.io/article/custom-handlers-avoiding-globals/
type werkerStreamer struct {
	conf      *WerkerConf
	out       storage.BuildOut
	sum       storage.BuildSum
	buildInfo map[string]*buildDatum
	consul    *consulet.Consulet
}

type buildDatum struct {
	sync.Mutex
	buildData [][]byte
	done      bool
}

func (b *buildDatum) GetData() [][]byte{
	return b.buildData
}

func (b *buildDatum) CheckDone() bool {
	return b.done
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
		if ok {
			err := streamer.StreamFromArray(buildInfo, stream, ocelog.Log().Debug)
			if err != nil {
				ocelog.IncludeErrField(err).Error("could not stream from array!")
			}
			ocelog.Log().Debug("streamed build data from array")
		} else {
			stream.SendError([]byte("did not find hash in current streaming data and the build was not marked as done"))
		}
	}
	defer stream.Finish(done)
}


// processTransport deals with adding info to consul, and calling writeInfoChanToInMemMap
func processTransport(transport *Transport, appCtx *werkerStreamer) {
	// question: does this support unicode?
	if err := rt.Register(appCtx.consul, transport.Hash, appCtx.conf.RegisterIP, appCtx.conf.grpcPort, appCtx.conf.ServicePort); err != nil {
		ocelog.IncludeErrField(err).Error("could not register with consul")
	} else {
		ocelog.Log().Infof("registered ip %s running build %s with consul", appCtx.conf.RegisterIP, transport.Hash)
	}
	writeInfoChanToInMemMap(transport, appCtx)
	// get rid of hash from cache, set build done in consul
	if err := rt.SetBuildDone(appCtx.consul, transport.Hash); err != nil {
		ocelog.IncludeErrField(err).Error("could not set build done")
	}
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
func writeInfoChanToInMemMap(transport *Transport, appCtx *werkerStreamer) {
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
		Output: bytes.Join(build.buildData, []byte("\n")),
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

func cacheProcessor(transpo chan *Transport, appCtx *werkerStreamer) {
	for i := range transpo {
		ocelog.Log().Debugf("adding info channel for hash %s to map for streaming access.", i.Hash)
		go processTransport(i, appCtx)
	}
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	//pwd, _ := os.Getwd()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("no caller???? ")
	}
	fmt.Println("FILENAME ",  path.Dir(filename) + "/test.html")
	http.ServeFile(w, r, path.Dir(filename) + "/test.html")
}

func getWerkerStreamer(conf *WerkerConf, store storage.OcelotStorage) *werkerStreamer {
	werkerConsul, err := consulet.Default()
	if err != nil {
		ocelog.IncludeErrField(err)
	}
	werkerStreamer := &werkerStreamer{
		conf:      conf,
		out:       store,
		sum:       store,
		buildInfo: make(map[string]*buildDatum),
		consul:    werkerConsul}
	return werkerStreamer
}

//ServeMe will start HTTP Server as needed for streaming build output by hash
func ServeMe(transportChan chan *Transport, conf *WerkerConf, store storage.OcelotStorage) {
	// todo: defer a recovery here
	werkStream := getWerkerStreamer(conf, store)
	ocelog.Log().Debug("saving build info channels to in memory map")
	go cacheProcessor(transportChan, werkStream)

	ocelog.Log().Info("serving websocket on port: ", conf.ServicePort)
	muxi := mux.NewRouter()
	muxi.Handle("/ws/builds/{hash}", &ocenet.AppContextHandler{werkStream, stream}).Methods("GET")
	muxi.HandleFunc("/builds/{hash}", serveHome).Methods("GET")

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

	ocelog.Log().Info("serving grpc streams of build data on port: ", conf.grpcPort)
	con, err := net.Listen("tcp", ":"+conf.grpcPort)
	if err != nil {
		ocelog.Log().Fatal("womp womp")
	}
	grpcServer := grpc.NewServer()
	protobuf.RegisterBuildServer(grpcServer, werkStream)
	go grpcServer.Serve(con)

}

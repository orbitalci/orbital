package werker

import (
	consulet "bitbucket.org/level11consulting/go-til/consul"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"bitbucket.org/level11consulting/ocelot/werker/protobuf"
	"bufio"
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"time"
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
	go pumpBundle(stream, w, request.Hash, pumpDone)
	<-pumpDone
	return nil
}

func stream(ctx interface{}, w http.ResponseWriter, r *http.Request) {
	a := ctx.(*werkerStreamer)
	vars := mux.Vars(r)
	hash := vars["hash"]
	ocelog.Log().Debug(hash)
	ws, err := upgrader.Upgrade(w, r, nil)
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

func writeStreamError(stream interface{}, description []byte) {
	switch strm := stream.(type) {
	case ocenet.WebsocketEy:
		writeWSError(strm, description)
	case protobuf.Build_BuildInfoServer:
		writeGrpcError(strm, description)
	}
}

func writeWSError(ws ocenet.WebsocketEy, description []byte) {
	ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	ws.WriteMessage(websocket.TextMessage, []byte("ERROR!\n"))
	ws.WriteMessage(websocket.TextMessage, description)
	ws.Close()
}

// writeWSError writes ERROR to the web socket along with a description and closes the web socket connection
func writeGrpcError(stream protobuf.Build_BuildInfoServer, description []byte) {
	stream.Send(&protobuf.Response{OutputLine: fmt.Sprintf("ERROR!\n %s", description)})
}

// pumpBundle writes build data to web socket
func pumpBundle(stream interface{}, appCtx *werkerStreamer, hash string, done chan int) {
	// determine whether to get from storage or off infoReader
	//if appCtx.CheckIfBuildDone(hash) {
	//	ocelog.Log().Debugf("build %s is done, getting from appCtx", hash)
	//	err := streamFromStorage(appCtx, hash, stream)
	//	if err != nil {
	//		ocelog.IncludeErrField(err).Error("error retrieving from storage")
	//	}
	//} else {
	//	ocelog.Log().Debug("pumping info array data to web socket")
		buildInfo, ok := appCtx.buildInfo[hash]
		if ok {
			err := streamFromArray(buildInfo, stream)
			if err != nil {
				ocelog.IncludeErrField(err).Error("could not stream from array!")
			}
			ocelog.Log().Debug("streamed build data from array")
		} else {
			writeStreamError(stream, []byte("did not find hash in current streaming data and the build was not marked as done"))
		}
	//}
	//defer cleanUpStream(stream, done)
}

func cleanUpStream(stream interface{}, done chan int) {
	switch strm := stream.(type) {
	case ocenet.WebsocketEy:
		close(done)
		strm.SetWriteDeadline(time.Now().Add(10 * time.Second))
		strm.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		strm.Close()
	}
}

// streamFromStorage gets the buildInfo data from storage and writes the lines to the websocket connection
func streamFromStorage(appCtx *werkerStreamer, hash string, stream interface{}) error {
	bytez, err := appCtx.storage.Retrieve(hash)
	if err != nil {
		writeStreamError(stream, []byte("could not retrieve persisted build data"))
		return err
	}
	reader := bytes.NewReader(bytez)
	s := bufio.NewScanner(reader)
	// write to web socket
	for s.Scan() {
		if err := send(stream, s.Bytes()); err != nil {
			ocelog.IncludeErrField(err).Error("could not write to stream")
			return err
		}
	}
	return s.Err()
}

// streamFromArray writes the buildData slice to a web socket. it keeps track of where the index is that it has
// previously read and waits for more data on the buildData slice until the buildInfo done flag is set to true,
// at which point it cancels out
func streamFromArray(buildInfo *buildDatum, stream interface{}) (err error) {
	var index int
	var previousIndex int
	for {
		time.Sleep(100) // todo: set polling to be configurable
		fullArrayStreamed := len(buildInfo.buildData) == index
		if buildInfo.done && fullArrayStreamed {
			ocelog.Log().Debug("done streaming from array")
			return nil
		}
		// if no new data has been sent, don't even try
		if fullArrayStreamed {
			continue
		}
		buildData := buildInfo.buildData[index:]
		ind, err := iterateOverBuildData(buildData, stream)
		previousIndex = index
		index += ind
		ocelog.Log().WithField("lines_sent", ind).WithField("index", index).WithField("previousIndex", previousIndex).Debug()
		if err != nil {
			return err
		}
	}

}

func send(conn interface{}, data []byte) error {
	switch ci := conn.(type) {
	case ocenet.WebsocketEy:
		// todo: figure out why this breaks shit if the time is too long (but not longer than the timeout???)
		//ci.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := ci.WriteMessage(websocket.TextMessage, data); err != nil {
			ocelog.IncludeErrField(err).Error("could not write to web socket")
			ci.Close()
			return err
		}
	case protobuf.Build_BuildInfoServer:
		if err := ci.Send(&protobuf.Response{OutputLine: string(data)}); err != nil {
			ocelog.IncludeErrField(err).Error("could not write to grpc stream")
			return err
		}
	default:
		fmt.Printf("CONNECTION IS WTF??? %v", ci)
	}
	return nil
}

func iterateOverBuildData(data [][]byte, stream interface{}) (int, error) {
	var index int
	for ind, dataLine := range data {
		if err := send(stream, dataLine); err != nil {
			return ind, err
		}
		// adding the number of lines added to index so streamFromArray knows where to start on the next pass
		index = ind + 1
	}
	return index, nil
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

func serveHome(w http.ResponseWriter, r *http.Request) {
	//TODO: need to change this to not use marianne's shit dummy
	http.ServeFile(w, r, "/Users/mariannefeng/go/src/bitbucket.org/level11consulting/ocelot/cmd/werker/test.html")
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
	//TODO: add new line here to serve everything out of test-fixtures
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

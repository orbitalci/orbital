package streaminglogs

import (
	"bytes"
	_ "net/http/pprof"
	"sync"
	"time"

	"github.com/level11consulting/ocelot/build/buildmonitor"
	"github.com/level11consulting/ocelot/client/runtime"
	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/storage"
	"github.com/shankj3/go-til/consul"
	ocelog "github.com/shankj3/go-til/log"
)

type buildDatum struct {
	sync.Mutex
	buildData [][]byte
	done      bool
}

func (b *buildDatum) Append(line []byte) {
	b.Lock()
	defer b.Unlock()
	b.buildData = append(b.buildData, line)
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

func GetStreamPack(store storage.OcelotStorage, consulet *consul.Consulet) *StreamPack {
	return &StreamPack{
		Consul:    consulet,
		Store:     store,
		BuildInfo: make(map[string]*buildDatum),
	}
}

// StreamPack holds all the connections that the grpc streamer needs to check if a build is active, and eventually
//   save the log data to the store
type StreamPack struct {
	Consul    *consul.Consulet
	Store     storage.OcelotStorage
	BuildInfo map[string]*buildDatum
}

// pumpBundle writes build data to web socket
func (sp *StreamPack) PumpBundle(stream Streamable, hash string, done chan int) {
	defer func() {
		if r := recover(); r != nil {
			ocelog.Log().WithField("recover", r).Error("recovered from a panic in pumpBundle!!")
		}
	}()
	// determine whether to get from out or off infoReader
	if runtime.CheckIfBuildDone(sp.Consul, sp.Store, hash) {
		ocelog.Log().Debugf("build %s is done, getting from appCtx", hash)
		latestSummary, err := sp.Store.RetrieveLatestSum(hash)
		if err != nil {
			ocelog.IncludeErrField(err).Error("could not get latest build from storage")
		} else {
			err = StreamFromStorage(sp.Store, stream, latestSummary.BuildId)
			if err != nil {
				ocelog.IncludeErrField(err).Error("error retrieving from storage")
			}
		}
	} else {
		ocelog.Log().Debug("pumping info array data to web socket")
		buildInfo, ok := sp.BuildInfo[hash]
		ocelog.Log().Debug("length of array to stream is %d", len(buildInfo.GetData()))
		if ok {
			err := StreamFromArray(buildInfo, stream, ocelog.Log())
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
func (sp *StreamPack) processTransport(transport *models.Transport) {
	sp.writeInfoChanToInMemMap(transport)
	// get rid of hash from cache, set build done in consul
	//if err := rt.SetBuildDone(appCtx.consul, transport.Hash); err != nil {
	//	ocelog.IncludeErrField(err).Error("could not set build done")
	//}
	ocelog.Log().Debugf("removing hash %s from readerCache, channelDict, and consul", transport.Hash)
	delete(sp.BuildInfo, transport.Hash)
	if err := buildmonitor.Delete(sp.Consul, transport.Hash); err != nil {
		ocelog.IncludeErrField(err).Error("could not recursively delete values from consul")
	}

}

// writeInfoChanToInMemMap is what processes transport objects that come from the transport channel (objects get created
// 	and sent when a new build is pulled off the queue).
// 	the info channel is written to an array which is put in a map in the appCtx along with a done channel so
//  there is a way to see when the array will not be written to anymore
//  when the info channel is closed and the loop finishes, all the data is written to the out defined in the
//  appCtx, the done flag is written to consul, and the array is removed from the map
func (sp *StreamPack) writeInfoChanToInMemMap(transport *models.Transport) {
	var dataSlice [][]byte
	build := &buildDatum{buildData: dataSlice, done: false}
	sp.BuildInfo[transport.Hash] = build
	ocelog.Log().Debugf("writing infochan data for %s", transport.Hash)
	for i := range transport.InfoChan {
		build.Append(i)
		time.Sleep(time.Millisecond)
	}
	ocelog.Log().Debug("done with build ", transport.Hash)
	out := &models.BuildOutput{
		BuildId: transport.DbId,
		Output:  bytes.Join(build.buildData, []byte("\n")),
	}
	//ocelog.Log().Debug(string(len(out.Output)))
	err := sp.Store.AddOut(out)
	// even if it didn't store properly, we need to set the build in the map as "done" so
	// that the streams that connect when the build is still happening know to close the connection
	build.done = true
	if err != nil {
		ocelog.IncludeErrField(err).Error("could not store build data to storage")
	}
}

func (sp *StreamPack) ListenTransport(transpo chan *models.Transport) {
	for i := range transpo {
		ocelog.Log().Debugf("adding info channel for hash %s to map for streaming access.", i.Hash)
		go sp.processTransport(i)
	}
}

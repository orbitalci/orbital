package streamer

import (
	_ "net/http/pprof"

	ocelog "bitbucket.org/level11consulting/go-til/log"

	rt "bitbucket.org/level11consulting/ocelot/build"
	"bitbucket.org/level11consulting/ocelot/models"

	"bytes"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)
// TODO: half of this belongs in router

var (
	upgrader = websocket.Upgrader{}
)


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


// pumpBundle writes build data to web socket
func PumpBundle(stream Streamable, appCtx *WerkerContext, hash string, done chan int) {
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
func processTransport(transport *models.Transport, appCtx *WerkerContext) {
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
func writeInfoChanToInMemMap(transport *models.Transport, appCtx *WerkerContext) {
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

func ListenTransport(transpo chan *models.Transport, appCtx *WerkerContext) {
	for i := range transpo {
		ocelog.Log().Debugf("adding info channel for hash %s to map for streaming access.", i.Hash)
		go processTransport(i, appCtx)
	}
}

func ListenBuilds(buildsChan chan *models.BuildContext, appCtx *WerkerContext, mapLock sync.Mutex) {
	for newBuild := range buildsChan {
		mapLock.Lock()
		ocelog.Log().Debugf("got new build context for %s", newBuild.Hash)
		appCtx.BuildContexts[newBuild.Hash] = newBuild
		mapLock.Unlock()
		go contextCleanup(newBuild, appCtx, mapLock)
	}
}

func contextCleanup(buildCtx *models.BuildContext, appCtx *WerkerContext, mapLock sync.Mutex) {
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


package werker

import (
	"bitbucket.org/level11consulting/ocelot/old/werker/protobuf"
	"fmt"
	"github.com/pkg/errors"
	"context"
	"bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/ocelot/util/buildruntime"
	"bitbucket.org/level11consulting/ocelot/newocy/build/cleaner"
)

//embeds the werkerappcontext so we can stream + access active builds
type WerkerServer struct {
	*WerkerContext
	cleaner.Cleaner
}

//streams logs for an active build
func (w *WerkerServer) BuildInfo(request *protobuf.Request, stream protobuf.Build_BuildInfoServer) error {
	stream.Send(wrap(request.Hash))
	stream.Send(wrap(w.Conf.WerkerName))
	stream.Send(wrap(w.Conf.RegisterIP))
	pumpDone := make(chan int)
	streamable := &protobuf.BuildStreamableServer{Server: stream}
	go pumpBundle(streamable, w.WerkerContext, request.Hash, pumpDone)
	<-pumpDone
	return nil
}

//handles build kills
func (w *WerkerServer) KillHash(request *protobuf.Request, stream protobuf.Build_KillHashServer) error {
	stream.Send(wrap(fmt.Sprintf("Checking active builds for %s...", request.Hash)))
	build, ok := w.BuildContexts[request.Hash]; if ok {
		stream.Send(wrap(fmt.Sprintf("An active build was found for %s, attempting to cancel...", request.Hash)))
		build.CancelFunc()

		// remove container
		stream.Send(wrap("Performing build cleanup..."))

		hashes, err := buildruntime.GetHashRuntimesByWerker(w.consul, w.Conf.WerkerUuid.String())
		if err != nil {
			log.IncludeErrField(err).Error("unable to retrieve active builds from consul")
			return err
		}
		build := hashes[request.Hash]
		if len(build.DockerUuid) > 0 {
			w.Cleanup(context.Background(), build.DockerUuid, nil)
			stream.Send(wrap(fmt.Sprintf("Successfully killed build for %s \u2713", request.Hash)))
		} else {
			stream.Send(wrap("Wow you killed your build before it even got to the setup stage??"))
		}
		if err = buildruntime.Delete(w.consul, request.Hash); err != nil {
			log.IncludeErrField(err).Error("couldn't delete out of consul")
			return errors.New("Couldn't delete build out of consul. Your build was killed, but cleanup didn't go as planned. Error: " + err.Error())
		}

		return nil
	}
	return errors.New(fmt.Sprintf("No active build was found for %s", request.Hash))
}

func NewWerkerServer(werkerCtx *WerkerContext) protobuf.BuildServer {
	werkerServer := &WerkerServer{
		WerkerContext: werkerCtx,
		Cleaner: cleaner.GetNewCleaner(werkerCtx.Conf.WerkerType),
	}
	return werkerServer
}


func wrap(textToSend string) *protobuf.Response {
	return &protobuf.Response{
		OutputLine: textToSend,
	}
}
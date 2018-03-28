package werker

import (
	"bitbucket.org/level11consulting/ocelot/werker/protobuf"
	"fmt"
	"github.com/pkg/errors"
)

//embeds the werkerappcontext so we can stream + access active builds
type WerkerServer struct {
	*WerkerContext
}

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

func (w *WerkerServer) KillHash(request *protobuf.Request, stream protobuf.Build_KillHashServer) error {
	stream.Send(wrap(fmt.Sprintf("Checking active builds for %s...", request.Hash)))
	build, ok := w.BuildContexts[request.Hash]; if ok {
		stream.Send(wrap(fmt.Sprintf("An active build was found for %s, attempting to cancel...", request.Hash)))
		build.CancelFunc()
		return nil
	}
	return errors.New(fmt.Sprintf("No active build was found for %s", request.Hash))
}

func NewWerkerServer(werkerCtx *WerkerContext) protobuf.BuildServer {
	werkerServer := &WerkerServer{
		WerkerContext: werkerCtx,
	}
	return werkerServer
}

func wrap(textToSend string) *protobuf.Response {
	return &protobuf.Response{
		OutputLine: textToSend,
	}
}
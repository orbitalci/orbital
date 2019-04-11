package werker

import (
	"context"
	"errors"
	"fmt"

	"github.com/shankj3/go-til/log"
	rt "github.com/level11consulting/ocelot/build"
	"github.com/level11consulting/ocelot/build/cleaner"
	"github.com/level11consulting/ocelot/build/streaminglogs"
	"github.com/level11consulting/ocelot/build/valet"
	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/models/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//WerkerServer embeds the werkerappcontext so we can stream + access active builds
type WerkerServer struct {
	*WerkerContext
	cleaner.Cleaner
}

//BuildInfo streams logs for an active build
func (w *WerkerServer) BuildInfo(request *pb.Request, stream pb.Build_BuildInfoServer) error {
	stream.Send(wrap(request.Hash))
	//stream.Send(wrap(w.Conf.WerkerName))
	//stream.Send(wrap(w.Conf.RegisterIP))
	pumpDone := make(chan int)
	streamable := &streaminglogs.BuildStreamableServer{Build_BuildInfoServer: stream}
	go w.streamPack.PumpBundle(streamable, request.Hash, pumpDone)
	<-pumpDone
	return nil
}

//KillHash handles build kills
func (w *WerkerServer) KillHash(request *pb.Request, stream pb.Build_KillHashServer) error {
	stream.Send(wrap(fmt.Sprintf("Checking active builds for %s...", request.Hash)))
	if err := w.killValet.Kill(request.Hash); err == nil {
		stream.Send(wrap(fmt.Sprintf("An active build was found for %s, attempting to cancel...", request.Hash)))

		// remove container
		stream.Send(wrap("Performing build cleanup..."))

		hashes, err := rt.GetHashRuntimesByWerker(w.consul, w.Uuid.String())
		if err != nil {
			log.IncludeErrField(err).Error("unable to retrieve active builds from consul")
			return status.Error(codes.Internal, err.Error())
		}
		build, ok := hashes[request.Hash]
		if ok {
			// this should be handled by valet, but docker doesn't handle killing a build that is expecting stdin (ie a request for for accepting integrity of an ssh key)
			// fixme: figure our how to be able to do this, because just cancelling in the context in the stdin scenario doesn't trigger docker api to kill the container
			err = w.Cleanup(context.Background(), build.DockerUuid, nil)
			stream.Send(wrap(fmt.Sprintf("Successfully killed build for %s %s", request.Hash, models.CHECKMARK)))
		} else {
			stream.Send(wrap("Wow you killed your build before it even got to the setup stage??"))
		}
		if err = valet.Delete(w.consul, request.Hash); err != nil {
			log.IncludeErrField(err).Error("couldn't delete out of consul")
			return errors.New("Couldn't delete build out of consul. Your build was killed, but cleanup didn't go as planned. Error: " + err.Error())
		}

		return nil
	}
	return status.Error(codes.NotFound, fmt.Sprintf("No active build was found for %s", request.Hash))
}

func NewWerkerServer(werkerCtx *WerkerContext) pb.BuildServer {
	werkerServer := &WerkerServer{
		WerkerContext: werkerCtx,
		Cleaner:       cleaner.GetNewCleaner(werkerCtx.WerkerType, werkerCtx.Ssh),
	}
	return werkerServer
}

func wrap(textToSend string) *pb.Response {
	return &pb.Response{
		OutputLine: textToSend,
	}
}

package werker

import (
	"bitbucket.org/level11consulting/ocelot/werker/protobuf"
	"fmt"
	"github.com/pkg/errors"
	"context"
	"github.com/docker/docker/client"
	"bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/ocelot/util/buildruntime"
	"github.com/docker/docker/api/types"
)

//embeds the werkerappcontext so we can stream + access active builds
type WerkerServer struct {
	*WerkerContext
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

		ctx := context.Background()
		cli, err := client.NewEnvClient()
		if err != nil {
			log.IncludeErrField(err).Error("unable to get docker client?? ")
			return err
		}

		hashes, err := buildruntime.GetHashRuntimesByWerker(w.consul, w.Conf.WerkerUuid.String())
		if err != nil {
			log.IncludeErrField(err).Error("unable to retrieve active builds from consul")
			return err
		}
		build := hashes[request.Hash]
		if len(build.DockerUuid) > 0 {
			//tODO: tHIS IS COPIED FROM VALET VERBATIM - EXTRACT?
			if err := cli.ContainerKill(ctx, build.DockerUuid, "SIGKILL"); err != nil {
				log.IncludeErrField(err).WithField("dockerId", build.DockerUuid).Error("could not kill container")
			} else {
				log.Log().WithField("dockerId", build.DockerUuid).Info("killed container")
			}

			// even if ther is an error with containerKill, it might be from the container already exiting (ie bad ocelot.yml). so still try to remove.
			log.Log().WithField("dockerId", build.DockerUuid).Info("removing")
			if err := cli.ContainerRemove(ctx, build.DockerUuid, types.ContainerRemoveOptions{}); err != nil {
				log.IncludeErrField(err).WithField("dockerId", build.DockerUuid).Error("could not rm container")
			} else {
				log.Log().WithField("dockerId", build.DockerUuid).Info("removed container")
			}

			cli.Close()
			stream.Send(wrap(fmt.Sprintf("Successfully killed build for %s \u2713", request.Hash)))
		} else {
			stream.Send(wrap("Wow you killed your build before it even got to the setup stage??"))
		}

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
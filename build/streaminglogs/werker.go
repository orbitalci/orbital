package streaminglogs

import (
	"github.com/level11consulting/ocelot/models/pb"
)

// mod for buildBuildInfoServer so that it plays nice with streamer package

type BuildStreamableServer struct {
	pb.Build_BuildInfoServer
}

func (x *BuildStreamableServer) SendIt(data []byte) error {

	resp := &pb.Response{OutputLine: string(data)}
	return x.Send(resp)
}

func (x *BuildStreamableServer) SendError(errorDesc []byte) {
	x.Send(&pb.Response{OutputLine: "Error!"})
	x.Send(&pb.Response{OutputLine: string(errorDesc)})
}

func (x *BuildStreamableServer) Finish(done chan int) {
	close(done)
}

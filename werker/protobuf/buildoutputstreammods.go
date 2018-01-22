package protobuf

// mod for buildBuildInfoServer so that it plays nice with streamer package

type BuildStreamableServer struct {
	Server Build_BuildInfoServer
}

func (x *BuildStreamableServer) SendIt(data []byte) error {
	resp := &Response{OutputLine: string(data)}
	return x.Server.Send(resp)
}

func (x *BuildStreamableServer) SendError(errorDesc []byte) {
	x.Server.Send(&Response{OutputLine: "Error!"})
	x.Server.Send(&Response{OutputLine: string(errorDesc)})
}

func (x *BuildStreamableServer) Finish(done chan int) {
	close(done)
}

package protobuf

// mod for buildBuildInfoServer so that it plays nice with streamer package

type BuildStreamableServer struct {
	Server Build_BuildInfoServer
}

func (x *BuildStreamableServer) SendIt(data []byte) error {
	return x.Server.Send(&Response{OutputLine: string(data)})
}

func (x *BuildStreamableServer) SendError(errorDesc []byte) {
	x.Server.Send(&Response{OutputLine: "Error!"})
	x.Server.Send(&Response{OutputLine: string(errorDesc)})
}

func (x *BuildStreamableServer) Finish() {}

package streamer


type StreamArray interface {
	GetData() [][]byte
	CheckDone() bool
}

type Streamable interface {
	SendIt(data []byte) error
	SendError(errorDesc []byte)
	Finish(chan int)
}

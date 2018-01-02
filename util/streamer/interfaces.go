package streamer


type StreamArray interface {
	GetData() [][]byte
	CheckDone() bool
	Lock()
	Unlock()
}

type Streamable interface {
	SendIt(data []byte) error
	SendError(errorDesc []byte)
	Finish(chan int)
}

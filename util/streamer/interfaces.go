package streamer

import "sync"

type StreamArray interface {
	GetData() [][]byte
	CheckDone() bool
	sync.Locker
}

type Streamable interface {
	SendIt(data []byte) error
	SendError(errorDesc []byte)
	Finish(chan int)
}

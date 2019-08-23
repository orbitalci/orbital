package streaminglogs

import "sync"

type StreamArray interface {
	GetData() [][]byte
	CheckDone() bool
	Append(line []byte)
	sync.Locker
}

type Streamable interface {
	SendIt(data []byte) error
	SendError(errorDesc []byte)
	Finish(chan int)
}

type Loggy interface {
	Debug(args ...interface{})
	Error(args ...interface{})
	Info(args ...interface{})
}

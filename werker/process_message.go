package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/shankj3/ocelot/protos/out"
	"github.com/shankj3/ocelot/util/nsqpb"
	"github.com/shankj3/ocelot/util/ocelog"
)

type WorkerMsgHandler struct {
	topic    string
	infochan chan string
}

// NewWorkerMsgHandler instantiates WorkerMsgHandler with the topic name
// and sets the infochan. atomic.
func NewWorkerMsgHandler(topic string) WorkerMsgHandler {
	return WorkerMsgHandler{
		topic: topic,
		infochan: make(chan string),
	}
}

func (w WorkerMsgHandler) UnmarshalAndProcess(msg []byte) error {
	unmarshalobj := nsqpb.TopicsUnmarshalObj(w.topic)
	//ocelog.Log().Debug("INSIDE UNMARSHAL AND PROCESS! isn't that nice!")
	if err := proto.Unmarshal(msg, unmarshalobj); err != nil {
		ocelog.IncludeErrField(err).Warning("unmarshal error")
		return err
	}
	w.Build(unmarshalobj)
	fmt.Println("receiving std out from channel: ")
	for i := range w.infochan {
		fmt.Println(i)
	}

	return nil
}

func (w *WorkerMsgHandler) Build(psg proto.Message) {
	switch v := psg.(type) {
	case *protos.PRBuildBundle:
		w.runPRBundle(v)
	case *protos.PushBuildBundle:
		w.runPushBundle(v)
	default:
		fmt.Println("why is there no timeeeeeeeeeeeeeeeeeee ", v)
	}
}

func (w *WorkerMsgHandler) runPushBundle(bund *protos.PushBuildBundle) {
	// run push bundle.
	//fmt.Println(bund.PushData.Repository.FullName)
	w.infochan <- bund.PushData.Repository.FullName
	w.infochan <- bund.PushData.Repository.Owner.Username
	w.infochan <- "push requeeeeeeeest"
	w.infochan <- "this could be some delightful std out from builds! huzzah!"
	close(w.infochan)
}

func (w *WorkerMsgHandler) runPRBundle(bund *protos.PRBuildBundle) {
	// run pr bundle.
	//fmt.Println("WOW I MADE IT ALL THE WAY TO RUN PR BUNDLE!")
	w.infochan <- bund.PrData.Repository.FullName
	w.infochan <- "pulllll reqquuuueeeeeeeeeest!"
	close(w.infochan)
}
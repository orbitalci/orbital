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
// was going to do *something* with info chan but it has to be reset every run of UnmarshalAndProcess
func NewWorkerMsgHandler(topic string) WorkerMsgHandler {
	return WorkerMsgHandler{
		topic: topic,
	}
}

func (w WorkerMsgHandler) UnmarshalAndProcess(msg []byte) error {
	ocelog.Log().Debug("unmarshaling build obj and processing")
	unmarshalobj := nsqpb.TopicsUnmarshalObj(w.topic)
	if err := proto.Unmarshal(msg, unmarshalobj); err != nil {
		ocelog.IncludeErrField(err).Warning("unmarshal error")
		return err
	}
	// channels get closed after the build finishes
	w.infochan = make(chan string)
	// set goroutine for watching for results and logging them (for rn)
	go w.watchForResults()
	// do the thing
	go w.build(unmarshalobj)
	return nil
}

func (w *WorkerMsgHandler) watchForResults() {
	ocelog.Log().Debug("oooooeee Watchin for results!!!")
	for i := range w.infochan {
		fmt.Println(i)
	}
	ocelog.Log().Debug("ooooeee finished watchin for results!!! recreating channel!f")
}


func (w *WorkerMsgHandler) build(psg proto.Message) {
	ocelog.Log().Debug("How exciting! I gonna build!")
	switch v := psg.(type) {
	case *protos.PRBuildBundle:
		w.runPRBundle(v)
	case *protos.PushBuildBundle:
		w.runPushBundle(v)
	default:
		fmt.Println("why is there no timeeeeeeeeeeeeeeeeeee ", v)
	}
	ocelog.Log().Debug("WOWOEE ZOWEE! finished building")
}

func (w *WorkerMsgHandler) runPushBundle(bund *protos.PushBuildBundle) {
	ocelog.Log().Debug("building building tasty tasty push bundle")
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
package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/shankj3/ocelot/protos/out"
	"github.com/shankj3/ocelot/util/nsqpb"
	"github.com/shankj3/ocelot/util/ocelog"
)

type WorkerMsgHandler struct {
	topic string
}

func (w WorkerMsgHandler) UnmarshalAndProcess(msg []byte) error {
	unmarshalobj := nsqpb.TopicsUnmarshalObj(w.topic)
	ocelog.Log().Debug("INSIDE UNMARSHAL AND PROCESS! isn't that nice!")
	if err := proto.Unmarshal(msg, unmarshalobj); err != nil {
		ocelog.IncludeErrField(err).Warning("unmarshal error")
		return err
	}
	w.Build(unmarshalobj)
	return nil
}

func (w *WorkerMsgHandler) Build(psg proto.Message) {
	switch v := psg.(type) {
	case *protos.PRBuildBundle:
		runPRBundle(v)
	case *protos.PushBuildBundle:
		runPushBundle(v)
	default:
		fmt.Println("why is there no timeeeeeeeeeeeeeeeeeee ", v)
	}
}

func runPushBundle(bund *protos.PushBuildBundle) {
	// run push bundle.
	fmt.Println("OOOOEEE I MADE IT ALL THE WAY TO RUN PUSH BUNDLE")
	fmt.Println(bund.PushData.Repository.FullName)
}

func runPRBundle(bund *protos.PRBuildBundle) {
	// run pr bundle.
	fmt.Println("WOW I MADE IT ALL THE WAY TO RUN PR BUNDLE!")
}
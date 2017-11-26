package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/shankj3/ocelot/protos/out"
	"github.com/shankj3/ocelot/util/nsqpb"
	"github.com/shankj3/ocelot/util/ocelog"
)

// Transport struct is for the Transport channel that will interact with the streaming side of the service
// to stream results back to the admin. It sends just enough to be unique, the hash that triggered the build
// and the InfoChan which the builder will write to.
type Transport struct {
	Hash       string
	InfoChan   chan []byte
}

type WorkerMsgHandler struct {
	topic    string
	werkConf *WerkerConf
	infochan chan []byte
	chanChan chan *Transport
}

// UnmarshalAndProcess is called by the nsq consumer to handle the build message
func (w WorkerMsgHandler) UnmarshalAndProcess(msg []byte) error {
	ocelog.Log().Debug("unmarshaling build obj and processing")
	unmarshalobj := nsqpb.TopicsUnmarshalObj(w.topic)
	if err := proto.Unmarshal(msg, unmarshalobj); err != nil {
		ocelog.IncludeErrField(err).Warning("unmarshal error")
		return err
	}
	// channels get closed after the build finishes
	w.infochan = make(chan []byte)
	// set goroutine for watching for results and logging them (for rn)
	// cant add go watchForResults here bc can't call method on interface until it's been cast properly.
	// do the thing
	go w.build(unmarshalobj)
	return nil
}

// watchForResults sends the *Transport object over the transport channel for stream functions to process
func (w *WorkerMsgHandler) WatchForResults(hash string) {
	ocelog.Log().Debugf("adding hash ( %s ) & infochan to transport channel", hash)
	transport := &Transport{Hash: hash, InfoChan: w.infochan,}
	w.chanChan <- transport
}

// build contains the logic for actually building. switches on type of proto message that was sent
// over the nsq queue
func (w *WorkerMsgHandler) build(psg nsqpb.BundleProtoMessage) {
	ocelog.Log().Debug("about to build")
	switch v := psg.(type) {
	case *protos.PRBuildBundle:
		ocelog.Log().Debug("hash build ", v.GetCheckoutHash())
		w.WatchForResults(v.GetCheckoutHash())
		w.werkConf.werkerProcessor.RunPRBundle(v, w.infochan)
	case *protos.PushBuildBundle:
		ocelog.Log().Debug("hash build: ", v.GetCheckoutHash())
		w.WatchForResults(v.GetCheckoutHash())
		w.werkConf.werkerProcessor.RunPushBundle(v, w.infochan)
	default:
		// todo: default handling
		fmt.Println("why is there no timeeeeeeeeeeeeeeeeeee ", v)
	}
	ocelog.Log().Debugf("finished building id %s", psg.GetCheckoutHash())
}
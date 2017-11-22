package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/shankj3/ocelot/protos/out"
	"github.com/shankj3/ocelot/util/nsqpb"
	"github.com/shankj3/ocelot/util/ocelog"
)

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

func (w *WorkerMsgHandler) watchForResults(hash string) {
	ocelog.Log().Debug("HASH!", hash)

	ocelog.Log().Debugf("adding hash (%s) & infochan to transport channel", hash)
	transport := &Transport{Hash: hash, InfoChan: w.infochan,}
	w.chanChan <- transport
	ocelog.Log().Debug("watchForResults thread started")
	//for i := range w.infochan {
	//	fmt.Println(i)
	//}
}


func (w *WorkerMsgHandler) build(psg nsqpb.BundleProtoMessage) {
	ocelog.Log().Debug("about to build")
	switch v := psg.(type) {
	case *protos.PRBuildBundle:
		ocelog.Log().Debug("hash build ", v.PrData.Pullrequest.Source.Commit.Hash)
		w.watchForResults(v.PrData.Pullrequest.Source.Commit.Hash)
		w.runPRBundle(v)
	case *protos.PushBuildBundle:
		ocelog.Log().Debug("hash build: ", v.PushData.Push.Changes[0].New.Target.Hash)
		w.watchForResults(v.PushData.Push.Changes[0].New.Target.Hash)
		w.runPushBundle(v)
	default:
		fmt.Println("why is there no timeeeeeeeeeeeeeeeeeee ", v)
	}
	ocelog.Log().Debugf("finished building id %s", psg.GetCheckoutHash())
}

// todo: dedupe all this shit. probably just switch earlier? idk seems weird.

func (w *WorkerMsgHandler) runPushBundle(bund *protos.PushBuildBundle) {
	switch w.werkConf.werkerType {
	case Docker:
		w.runDockerPushBundle(bund)
	case Kubernetes:
		w.runK8sPushBundle(bund)
	}
}

func (w *WorkerMsgHandler) runDockerPushBundle(bund *protos.PushBuildBundle) {
		ocelog.Log().Debug("building building tasty tasty push bundle")
		// run push bundle.
		//fmt.Println(bund.PushData.Repository.FullName)
		w.infochan <- []byte(bund.PushData.Repository.FullName)
		w.infochan <- []byte(bund.PushData.Repository.Owner.Username)
		w.infochan <- []byte("push requeeeeeeeest DOCKER!")
		w.infochan <- []byte("this could be some delightful std out from builds! huzzah! I'M RUNNING W/ DOCKER!")
		close(w.infochan)
}


func (w *WorkerMsgHandler) runK8sPushBundle(bund *protos.PushBuildBundle) {
	ocelog.Log().Debug("building building tasty tasty push bundle")
	// run push bundle.
	//fmt.Println(bund.PushData.Repository.FullName)
	w.infochan <- []byte(bund.PushData.Repository.FullName)
	w.infochan <- []byte(bund.PushData.Repository.Owner.Username)
	w.infochan <- []byte("push requeeeeeeeest KUBERNETES!")
	w.infochan <- []byte("this could be some delightful std out from builds! huzzah! I'M RUNNING W/ k8s!!")
	close(w.infochan)
}


func (w *WorkerMsgHandler) runPRBundle(bund *protos.PRBuildBundle) {
	// run pr bundle.
	//fmt.Println("WOW I MADE IT ALL THE WAY TO RUN PR BUNDLE!")
	switch w.werkConf.werkerType {
	case Docker:
		w.runDockerPRBundle(bund)
	case Kubernetes:
		w.runK8sPRBundle(bund)

	}
}

func (w *WorkerMsgHandler) runDockerPRBundle(bund *protos.PRBuildBundle) {
	w.infochan <- []byte(bund.PrData.Repository.FullName)
	w.infochan <- []byte("delightful! docker! love docker!")
	w.infochan <- []byte("dockeeerrr pulllll reqquuuueeeeeeeeeest!")
	close(w.infochan)
}


func (w *WorkerMsgHandler) runK8sPRBundle(bund *protos.PRBuildBundle) {
	w.infochan <- []byte(bund.PrData.Repository.FullName)
	w.infochan <- []byte("even better! love k8s! cool cool cool!!")
	w.infochan <- []byte("kubernetteeees pulllll reqquuuueeeeeeeeeest!")
	close(w.infochan)
}
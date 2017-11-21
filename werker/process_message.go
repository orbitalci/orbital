package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/shankj3/ocelot/protos/out"
	"github.com/shankj3/ocelot/util/nsqpb"
	"github.com/shankj3/ocelot/util/ocelog"
	"time"
)

var chanDict = NewCD()

type WorkerMsgHandler struct {
	topic    string
	werkConf *WerkerConf
	infochan chan string
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
	time.Sleep(0.5*time.Second)
	return nil
}

func (w *WorkerMsgHandler) watchForResults() {
	ocelog.Log().Debug("watchForResults thread started")
	for i := range w.infochan {
		fmt.Println(i)
	}
}


func (w *WorkerMsgHandler) build(psg nsqpb.BundleProtoMessage) {
	chanDict.CarefulPut(psg.GetCheckoutHash(), w.infochan)
	switch v := psg.(type) {
	case *protos.PRBuildBundle:
		w.runPRBundle(v)
	case *protos.PushBuildBundle:
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
		w.infochan <- bund.PushData.Repository.FullName
		w.infochan <- bund.PushData.Repository.Owner.Username
		w.infochan <- "push requeeeeeeeest DOCKER!"
		w.infochan <- "this could be some delightful std out from builds! huzzah! I'M RUNNING W/ DOCKER!"
		close(w.infochan)
}


func (w *WorkerMsgHandler) runK8sPushBundle(bund *protos.PushBuildBundle) {
	ocelog.Log().Debug("building building tasty tasty push bundle")
	// run push bundle.
	//fmt.Println(bund.PushData.Repository.FullName)
	w.infochan <- bund.PushData.Repository.FullName
	w.infochan <- bund.PushData.Repository.Owner.Username
	w.infochan <- "push requeeeeeeeest KUBERNETES!"
	w.infochan <- "this could be some delightful std out from builds! huzzah! I'M RUNNING W/ k8s!!"
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
	w.infochan <- bund.PrData.Repository.FullName
	w.infochan <- "delightful! docker! love docker!"
	w.infochan <- "dockeeerrr pulllll reqquuuueeeeeeeeeest!"
	close(w.infochan)
}


func (w *WorkerMsgHandler) runK8sPRBundle(bund *protos.PRBuildBundle) {
	w.infochan <- bund.PrData.Repository.FullName
	w.infochan <- "even better! love k8s! cool cool cool!!"
	w.infochan <- "kubernetteeees pulllll reqquuuueeeeeeeeeest!"
	close(w.infochan)
}
package main

import (
	"github.com/shankj3/ocelot/protos/out"
	"github.com/shankj3/ocelot/util/storage"
	"testing"
)

var (
	pushBundleInfoMsg = []byte("hit run push bundle")
	prBundleInfoMsg   = []byte("hit run pr bundle")
)

type testWerkerProcessor struct {
	t *testing.T
}

func (t *testWerkerProcessor) RunPushBundle(bund *protos.PushBuildBundle, infoChan chan []byte) {
	infoChan <- pushBundleInfoMsg
}

func (t *testWerkerProcessor) RunPRBundle(bund *protos.PRBuildBundle, infoChan chan []byte) {
	infoChan <- prBundleInfoMsg
}

func testGetConf() *WerkerConf {
	return &WerkerConf{
		servicePort: 	 "9090",
		werkerName:  	 "test agent",
		werkerType:  	 Docker,
		werkerProcessor: &testWerkerProcessor{},
		storage:		 &storage.FileBuildStorage{}, // todo: create test interface
		logLevel: 		 "info",
	}
}

// testGetWorkerMsgHandler returns a WorkerMsgHandler with a werker configuration that
// uses the mock processor
func testGetWorkerMsgHandler(t *testing.T, topic string) *WorkerMsgHandler {
	werkConf := testGetConf()
	werkConf.werkerProcessor = &testWerkerProcessor{t: t,}

	tunnel := make(chan *Transport)
	infor := make(chan []byte)
	wmh := &WorkerMsgHandler{
		topic: topic,
		werkConf: werkConf,
		infochan: infor,
		chanChan: tunnel,
	}
	// set werker processor to mock one
	return wmh
}

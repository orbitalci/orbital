package werker

import (
	d "bitbucket.org/level11consulting/go-til/deserialize"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"github.com/golang/protobuf/proto"
	b "bitbucket.org/level11consulting/ocelot/werker/builder"
)

// Transport struct is for the Transport channel that will interact with the streaming side of the service
// to stream results back to the admin. It sends just enough to be unique, the hash that triggered the build
// and the InfoChan which the builder will write to.
type Transport struct {
	Hash     string
	InfoChan chan []byte
}

type WorkerMsgHandler struct {
	Topic        string
	WerkConf     *WerkerConf
	infochan     chan []byte
	ChanChan     chan *Transport
	Deserializer d.Deserializer
}

// UnmarshalAndProcess is called by the nsq consumer to handle the build message
func (w WorkerMsgHandler) UnmarshalAndProcess(msg []byte) error {
	ocelog.Log().Debug("unmarshaling build obj and processing")
	werkerTask := &pb.WerkerTask{}
	if err := proto.Unmarshal(msg, werkerTask); err != nil {
		ocelog.IncludeErrField(err).Warning("unmarshal error")
		return err
	}
	// channels get closed after the build finishes
	w.infochan = make(chan []byte)
	// set goroutine for watching for results and logging them (for rn)
	// cant add go watchForResults here bc can't call method on interface until it's been cast properly.

	var builder b.Builder
	switch w.WerkConf.werkerType {
	case Docker:
		builder = b.NewDockerBuilder()
		//
		//case Kubernetes:
		//	builder = b.NewK8Builder()
	}
	go w.build(werkerTask, builder)
	return nil
}

// watchForResults sends the *Transport object over the transport channel for stream functions to process
func (w *WorkerMsgHandler) WatchForResults(hash string) {
	ocelog.Log().Debugf("adding hash ( %s ) & infochan to transport channel", hash)
	transport := &Transport{Hash: hash, InfoChan: w.infochan}
	w.ChanChan <- transport
}

//TODO: make this so that you only call NewEnvClient once
// build will call appropriate builder functions
func (w *WorkerMsgHandler) build(werk *pb.WerkerTask, builder b.Builder) {
	ocelog.Log().Debug("hash build ", werk.CheckoutHash)
	defer close(w.infochan)
	w.WatchForResults(werk.CheckoutHash)

	//TODO: do something with outputs here
	result := builder.Setup(w.infochan, werk.BuildConf.Image, werk.BuildConf.Env)
	if result.Status == b.FAIL {
		//WRITE TO DB
		return
	}

	for stageKey, stageVal := range werk.BuildConf.Stages {
		//build is special because we deploy with this
		if stageKey == "build" {
			builder.Build(w.infochan)
		}
		builder.Execute(stageKey, stageVal, w.infochan)
	}

	builder.Cleanup()

	ocelog.Log().Debugf("finished building id %s", werk.CheckoutHash)
}

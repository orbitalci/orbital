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
	go w.build(werkerTask)
	return nil
}

// watchForResults sends the *Transport object over the transport channel for stream functions to process
func (w *WorkerMsgHandler) WatchForResults(hash string) {
	ocelog.Log().Debugf("adding hash ( %s ) & infochan to transport channel", hash)
	transport := &Transport{Hash: hash, InfoChan: w.infochan}
	w.ChanChan <- transport
}

// build contains the logic for actually building. switches on type of proto message that was sent
// over the nsq queue
func (w *WorkerMsgHandler) build(werk *pb.WerkerTask) {
	ocelog.Log().Debug("hash build ", werk.CheckoutHash)
	w.WatchForResults(werk.CheckoutHash)

	var builder b.Builder

	switch w.WerkConf.werkerType {
	case Docker:
		builder = b.NewDockerBuilder()
	//
	//case Kubernetes:
	//	builder = b.NewK8Builder()
	}

	//TODO: do something with outputs here
	result := builder.Setup(w.infochan, werk.BuildConf.Image)
	ocelog.Log().Debug(result)
	//beforeResult := builder.Before(w.infochan)
	//buildResult := builder.Build(w.infochan)
	//afterResult := builder.After(w.infochan)
	//testResult := builder.Test(w.infochan)
	//deployResult := builder.Deploy(w.infochan)

	ocelog.Log().Debugf("finished building id %s", werk.CheckoutHash)
}

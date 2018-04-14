package listener


import (
	ocelog "bitbucket.org/level11consulting/go-til/log"
	pb "bitbucket.org/level11consulting/ocelot/old/protos"
	"bitbucket.org/level11consulting/ocelot/util/buildruntime"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/newocy/integrations/dockr"
	"bitbucket.org/level11consulting/ocelot/newocy/integrations/k8s"
	"bitbucket.org/level11consulting/ocelot/newocy/integrations/nexus"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"bitbucket.org/level11consulting/ocelot/util/storage/models"
	b "bitbucket.org/level11consulting/ocelot/old/werker/builder"
	"bitbucket.org/level11consulting/ocelot/newocy/build/valet"
	"fmt"
	"github.com/golang/protobuf/proto"
	"strings"

	//"runtime/debug"
	"context"
	"time"
	"bitbucket.org/level11consulting/ocelot/old/werker/config"
)

// Transport struct is for the Transport channel that will interact with the streaming side of the service
// to stream results back to the admin. It sends just enough to be unique, the hash that triggered the build
// and the InfoChan which the builder will write to.
type Transport struct {
	Hash     string
	InfoChan chan []byte
	DbId     int64
}

type BuildContext struct {
	Hash string
	Context context.Context
	CancelFunc func()
}

type WorkerMsgHandler struct {
	Topic           string
	WerkConf        *config.WerkerConf
	infochan        chan []byte
	StreamChan   chan *Transport
	BuildCtxChan chan *BuildContext
	Basher          *b.Basher
	Store           storage.OcelotStorage
	BuildValet   *valet.Valet

}

func NewWorkerMsgHandler(topic string, wc *config.WerkerConf, b *b.Basher, st storage.OcelotStorage, bv *valet.Valet, tunnel chan *Transport, buildChan chan *BuildContext) *WorkerMsgHandler {
	return &WorkerMsgHandler{
		Topic: 		topic,
		WerkConf: 	wc,
		Basher: 	b,
		Store: 		st,
		BuildValet: bv,
		StreamChan:   tunnel,
		BuildCtxChan: buildChan,
	}
}

// UnmarshalAndProcess is called by the nsq consumer to handle the build message
// It uses two channels to communicate with nsqpb, done and finish.
// the done channel is just sent at the end and is used in nsqpb to ensure that the queue is "Touch"ed at a
// set interval so that the message doesn't time out. The finish channel is for improper exits; ie panic recover
// or signal handling
// The nsqpb will call msg.Finish() when it receives on this channel.
func (w WorkerMsgHandler) UnmarshalAndProcess(msg []byte, done chan int, finish chan int) error {
	ocelog.Log().Debug("unmarshal-ing build obj and processing")
	werkerTask := &pb.WerkerTask{}
	if err := proto.Unmarshal(msg, werkerTask); err != nil {
		ocelog.IncludeErrField(err).Warning("unmarshal error")
		return err
	}
	if err := w.Store.StartBuild(werkerTask.Id); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't log start of build, returning")
		return err
	}
	// channels get closed after the build finishes
	w.infochan = make(chan []byte)
	ocelog.Log().Debug(fmt.Sprintf("INFO CHANNEL IS!!!!!  %v     MSGHANDLER IS!!!! %#v", w.infochan, w))
	// set goroutine for watching for results and logging them (for rn)
	// cant add go watchForResults here bc can't call method on interface until it's been cast properly.
	//
	var builder b.Builder
	switch w.WerkConf.WerkerType {
	case config.Docker:
		builder = b.NewDockerBuilder(w.Basher)
	default:
		builder = b.NewDockerBuilder(w.Basher)
	}
	w.MakeItSo(werkerTask, builder, finish, done)
	return nil
}


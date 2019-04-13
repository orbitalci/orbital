package werkerevent

import (
	"github.com/golang/protobuf/proto"
	builderinterface "github.com/level11consulting/ocelot/build/builder/interface"
	"github.com/level11consulting/ocelot/build/builder/runtime"
	"github.com/level11consulting/ocelot/build/builder/shell"
	"github.com/level11consulting/ocelot/build/builder/type/docker"
	"github.com/level11consulting/ocelot/build/builder/type/exec"
	"github.com/level11consulting/ocelot/build/builder/type/ssh"
	"github.com/level11consulting/ocelot/build/valet"
	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/server/config"
	"github.com/level11consulting/ocelot/storage"
	"github.com/prometheus/client_golang/prometheus"
	ocelog "github.com/shankj3/go-til/log"

	//"runtime/debug"
	"fmt"
)

var (
	recievedMsgs = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "ocelot_received_messages",
		Help: "number of messages received by node",
	})
)

func init() {
	prometheus.MustRegister(recievedMsgs)
}

type WorkerMsgHandler struct {
	*models.WerkerFacts
	Topic        string
	infochan     chan []byte
	StreamChan   chan *models.Transport
	BuildCtxChan chan *models.BuildContext
	Basher       *shell.Basher
	Store        storage.OcelotStorage
	BuildValet   *valet.Valet
	RemoteConfig config.CVRemoteConfig
}

func NewWorkerMsgHandler(topic string, facts *models.WerkerFacts, b *shell.Basher, st storage.OcelotStorage, bv *valet.Valet, rc config.CVRemoteConfig, tunnel chan *models.Transport, buildChan chan *models.BuildContext) *WorkerMsgHandler {
	return &WorkerMsgHandler{
		Topic:        topic,
		Basher:       b,
		Store:        st,
		BuildValet:   bv,
		StreamChan:   tunnel,
		BuildCtxChan: buildChan,
		RemoteConfig: rc,
		WerkerFacts:  facts,
	}
}

// UnmarshalAndProcess is called by the nsq consumer to handle the build message
// It uses two channels to communicate with nsqpb, done and finish.
// the done channel is just sent at the end and is used in nsqpb to ensure that the queue is "Touch"ed at a
// set interval so that the message doesn't time out. The finish channel is for improper exits; ie panic recover
// or signal handling
// The nsqpb will call msg.Finish() when it receives on this channel.
func (w WorkerMsgHandler) UnmarshalAndProcess(msg []byte, done chan int, finish chan int) error {
	recievedMsgs.Inc()
	var err error
	ocelog.Log().Debug("unmarshal-ing build obj and processing")
	werkerTask := &pb.WerkerTask{}
	if err = proto.Unmarshal(msg, werkerTask); err != nil {
		ocelog.IncludeErrField(err).Warning("unmarshal error")
		return err
	}
	if err = w.Store.StartBuild(werkerTask.Id); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't log start of build, returning")
		return err
	}
	// channels get closed after the build finishes
	w.infochan = make(chan []byte)
	ocelog.Log().Debug(fmt.Sprintf("INFO CHANNEL IS!!!!!  %v     MSGHANDLER IS!!!! %#v", w.infochan, w))
	// set goroutine for watching for results and logging them (for rn)
	// cant add go watchForResults here bc can't call method on interface until it's been cast properly.
	//
	var builder builderinterface.Builder
	switch w.WerkerType {
	case models.Docker:
		builder = docker.NewDockerBuilder(w.Basher)
	case models.SSH:
		builder, err = ssh.NewSSHBuilder(w.Basher, w.WerkerFacts)
		if err != nil {
			return err
		}
	case models.Exec:
		builder = exec.NewExecBuilder(w.Basher, w.WerkerFacts)
		if err != nil {
			return err
		}
	default:
		builder = docker.NewDockerBuilder(w.Basher)
	}
	launch := runtime.NewLauncher(w.WerkerFacts, w.RemoteConfig, w.StreamChan, w.BuildCtxChan, w.Basher, w.Store, w.BuildValet)
	launch.MakeItSo(werkerTask, builder, finish, done)
	return nil
}

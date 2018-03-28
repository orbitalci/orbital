package werker

import (
	ocelog "bitbucket.org/level11consulting/go-til/log"
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"bitbucket.org/level11consulting/ocelot/util/buildruntime"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"bitbucket.org/level11consulting/ocelot/util/storage/models"
	b "bitbucket.org/level11consulting/ocelot/werker/builder"
	"bitbucket.org/level11consulting/ocelot/werker/valet"
	"fmt"
	"github.com/golang/protobuf/proto"
	//"runtime/debug"
	"time"
)

// Transport struct is for the Transport channel that will interact with the streaming side of the service
// to stream results back to the admin. It sends just enough to be unique, the hash that triggered the build
// and the InfoChan which the builder will write to.
type Transport struct {
	Hash     string
	InfoChan chan []byte
	DbId     int64
}

type WorkerMsgHandler struct {
	Topic        string
	WerkConf     *WerkerConf
	infochan     chan []byte
	ChanChan     chan *Transport
	Basher       *b.Basher
	Store        storage.OcelotStorage
	BuildValet   *valet.Valet
}

//Topic:    	topic,
//WerkConf: 	conf,
//ChanChan: 	tunnel,
//Basher: 	basher,
//Store:  	store,
//BuildValet: bv,

func NewWorkerMsgHandler(topic string, wc *WerkerConf, b *b.Basher, st storage.OcelotStorage, bv *valet.Valet, tunnel chan *Transport) *WorkerMsgHandler {
	return &WorkerMsgHandler{
		Topic: 		topic,
		WerkConf: 	wc,
		Basher: 	b,
		Store: 		st,
		BuildValet: bv,
		ChanChan:   tunnel,
	}
}

// UnmarshalAndProcess is called by the nsq consumer to handle the build message
// It uses two channels to communicate with nsqpb, done and finish.
// the done channel is just sent at the end and is used in nsqpb to ensure that the queue is "Touch"ed at a
// set interval so that the message doesn't time out. The finish channel is for improper exits; ie panic recover
// or signal handling (TODO: figure out how to SIGNAL HANDLE)
// The nsqpb will call msg.Finish() when it receives on this channel.
func (w WorkerMsgHandler) UnmarshalAndProcess(msg []byte, done chan int, finish chan int) error {
	ocelog.Log().Debug("unmarshal-ing build obj and processing")
	werkerTask := &pb.WerkerTask{}
	if err := proto.Unmarshal(msg, werkerTask); err != nil {
		ocelog.IncludeErrField(err).Warning("unmarshal error")
		return err
	}
	// channels get closed after the build finishes
	w.infochan = make(chan []byte)
	ocelog.Log().Debug(fmt.Sprintf("INFO CHANNEL IS!!!!!  %v     MSGHANDLER IS!!!! %p", w.infochan, w))
	// set goroutine for watching for results and logging them (for rn)
	// cant add go watchForResults here bc can't call method on interface until it's been cast properly.
	//
	var builder b.Builder
	switch w.WerkConf.WerkerType {
	case Docker:
		builder = b.NewDockerBuilder(w.Basher)
	default:
		builder = b.NewDockerBuilder(w.Basher)
	}


	w.MakeItSo(werkerTask, builder, finish, done)
	return nil
}

// watchForResults sends the *Transport object over the transport channel for stream functions to process
func (w *WorkerMsgHandler) WatchForResults(hash string, dbId int64) {
	ocelog.Log().Debugf("adding hash ( %s ) & infochan to transport channel", hash)
	transport := &Transport{Hash: hash, InfoChan: w.infochan, DbId: dbId}
	w.ChanChan <- transport
}


//todo: build kill
// MakeItSo will call appropriate builder functions
func (w *WorkerMsgHandler) MakeItSo(werk *pb.WerkerTask, builder b.Builder, finish, done chan int) {
	ocelog.Log().Debug("hash build ", werk.CheckoutHash)

	w.BuildValet.RegisterDoneChan(werk.CheckoutHash, done)
	defer w.BuildValet.MakeItSoDed(finish)
	defer w.BuildValet.UnregisterDoneChan(werk.CheckoutHash)
	defer close(w.infochan)
	defer builder.Cleanup(w.infochan)

	w.WatchForResults(werk.CheckoutHash, werk.Id)

	consul := w.WerkConf.RemoteConfig.GetConsul()
	// if we can't register with consul, bail, just exit out. the maintainer will soon be pausing message flow anyway
	if err := w.BuildValet.StartBuild(consul, werk.CheckoutHash, werk.Id); err != nil {
		return
	}

	setupStart := time.Now()
	w.BuildValet.Reset("setup", werk.CheckoutHash)
	setupResult, uuid := builder.Setup(w.infochan, werk, w.WerkConf.RemoteConfig, w.WerkConf.ServicePort)
	setupDura := time.Now().Sub(setupStart)
	//defers are stacked, will be executed FILO
	if err := buildruntime.RegisterBuild(consul, w.WerkConf.WerkerUuid.String(), werk.CheckoutHash, uuid); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't register build")
		return

	}

	if err := storeStageToDb(w.Store, werk.Id, setupResult, setupStart, setupDura.Seconds()); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't store build output")
		return
	}

	if setupResult.Status == pb.StageResultVal_FAIL {
		errStr := "setup stage failed "
		if len(setupResult.Error) > 0 {
			errStr = errStr + setupResult.Error
		}
		ocelog.Log().Error(errStr)
		if err := w.Store.UpdateSum(true, setupDura.Seconds(), werk.Id); err != nil {
			ocelog.IncludeErrField(err).Error("couldn't update summary in database")
			return
		}
		return
	}
	fail := false
	start := time.Now()
	for _, stage := range werk.BuildConf.Stages {
		w.BuildValet.Reset(stage.Name, werk.CheckoutHash)
		stageStart := time.Now()
		stageResult := builder.Execute(stage, w.infochan, werk.CheckoutHash)
		ocelog.Log().WithField("hash", werk.CheckoutHash).Info("finished stage: ", stage.Name)
		if stageResult.Status == pb.StageResultVal_FAIL {
			fail = true
			stageDura := time.Now().Sub(stageStart)
			if err := storeStageToDb(w.Store, werk.Id, stageResult, stageStart, stageDura.Seconds()); err != nil {
				ocelog.IncludeErrField(err).Error("couldn't store build output")
				return
			}
			break
		}

		stageDura := time.Now().Sub(stageStart)
		if err := storeStageToDb(w.Store, werk.Id, stageResult, stageStart, stageDura.Seconds()); err != nil {
			ocelog.IncludeErrField(err).Error("couldn't store build output")
			return
		}

	}
	dura := time.Now().Sub(start)
	if err := w.Store.UpdateSum(fail, dura.Seconds(), werk.Id); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't update summary in database")
		return
	}

	ocelog.Log().Debugf("finished building id %s", werk.CheckoutHash)
}


//storeStageToDb is a helper function for storing stages to db - this runs on completion of every stage
func storeStageToDb(storage storage.OcelotStorage, buildId int64, stageResult *pb.Result, start time.Time, dur float64) error {
	err := storage.AddStageDetail(&models.StageResult{
		BuildId: buildId,
		Stage: stageResult.Stage,
		Status: int(stageResult.Status),
		Error: stageResult.Error,
		Messages: stageResult.Messages,
		StartTime: start,
		StageDuration: dur,
	})

	if err != nil {
		return err
	}

	return nil
}
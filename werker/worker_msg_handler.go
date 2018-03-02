package werker

import (
	d "bitbucket.org/level11consulting/go-til/deserialize"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"bitbucket.org/level11consulting/ocelot/util/buildruntime"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"bitbucket.org/level11consulting/ocelot/util/storage/models"
	b "bitbucket.org/level11consulting/ocelot/werker/builder"
	"github.com/golang/protobuf/proto"
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
	Deserializer d.Deserializer
	Basher	     *b.Basher
	Store        storage.OcelotStorage
}

// UnmarshalAndProcess is called by the nsq consumer to handle the build message
func (w WorkerMsgHandler) UnmarshalAndProcess(msg []byte, done chan int) error {
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
		builder = b.NewDockerBuilder(w.Basher)
	default:
		builder = b.NewDockerBuilder(w.Basher)
	}

	w.MakeItSo(werkerTask, builder)
	done <- 1
	return nil
}

// watchForResults sends the *Transport object over the transport channel for stream functions to process
func (w *WorkerMsgHandler) WatchForResults(hash string, dbId int64) {
	ocelog.Log().Debugf("adding hash ( %s ) & infochan to transport channel", hash)
	transport := &Transport{Hash: hash, InfoChan: w.infochan, DbId: dbId}
	w.ChanChan <- transport
}

// MakeItSo will call appropriate builder functions
func (w *WorkerMsgHandler) MakeItSo(werk *pb.WerkerTask, builder b.Builder) {
	ocelog.Log().Debug("hash build ", werk.CheckoutHash)

	defer close(w.infochan)
	defer builder.Cleanup(w.infochan)

	w.WatchForResults(werk.CheckoutHash, werk.Id)

	consul := w.WerkConf.RemoteConfig.GetConsul()
	if err := buildruntime.RegisterStartedBuild(consul, w.WerkConf.WerkerUuid.String(), werk.CheckoutHash); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't register build")
	}

	setupStart := time.Now()
	setupResult, uuid := builder.Setup(w.infochan, werk, w.WerkConf.RemoteConfig, w.WerkConf.ServicePort)
	setupDura := time.Now().Sub(setupStart)
	//defers are stacked, will be executed FILO
	if err := buildruntime.RegisterBuild(consul, w.WerkConf.WerkerUuid.String(), werk.CheckoutHash, uuid); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't register build")
	}

	if err := storeStageToDb(w.Store, werk.Id, setupResult, setupStart, setupDura.Seconds()); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't store build output")
	}

	if setupResult.Status == b.FAIL {
		errStr := "setup stage failed "
		if setupResult.Error != nil {
			errStr = errStr + setupResult.Error.Error()
		}
		ocelog.Log().Error(errStr)
		return
	}
	fail := false
	start := time.Now()
	for _, stage := range werk.BuildConf.Stages {
		stageStart := time.Now()
		stageResult := builder.Execute(stage, w.infochan, werk.CheckoutHash)
		ocelog.Log().WithField("hash", werk.CheckoutHash).Info("finished stage: ", stage.Name)
		if stageResult.Status == b.FAIL {
			fail = true
			stageDura := time.Now().Sub(stageStart)
			if err := storeStageToDb(w.Store, werk.Id, stageResult, stageStart, stageDura.Seconds()); err != nil {
				ocelog.IncludeErrField(err).Error("couldn't store build output")
			}
			break
		}

		stageDura := time.Now().Sub(stageStart)
		if err := storeStageToDb(w.Store, werk.Id, stageResult, stageStart, stageDura.Seconds()); err != nil {
			ocelog.IncludeErrField(err).Error("couldn't store build output")
		}

	}
	dura := time.Now().Sub(start)
	if err := w.Store.UpdateSum(fail, dura.Seconds(), werk.Id); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't update summary in database")
	}

	ocelog.Log().Debugf("finished building id %s", werk.CheckoutHash)
}


//storeStageToDb is a helper function for storing stages to db - this runs on completion of every stage
func storeStageToDb(storage storage.OcelotStorage, buildId int64, stageResult *b.Result, start time.Time, dur float64) error {
	var stageErr string

	//convert error to string for storing to db if exists
	if stageResult.Error != nil {
		stageErr = stageResult.Error.Error()
	}

	err := storage.AddStageDetail(&models.StageResult{
		BuildId: buildId,
		Stage: stageResult.Stage,
		Status: int(stageResult.Status),
		Error: stageErr,
		Messages: stageResult.Messages,
		StartTime: start,
		StageDuration: dur,
	})

	if err != nil {
		return err
	}

	return nil
}
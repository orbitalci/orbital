package werker

import (
	ocelog "bitbucket.org/level11consulting/go-til/log"
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"bitbucket.org/level11consulting/ocelot/util/buildruntime"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/integrations/dockr"
	"bitbucket.org/level11consulting/ocelot/util/integrations/k8s"
	"bitbucket.org/level11consulting/ocelot/util/integrations/nexus"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"bitbucket.org/level11consulting/ocelot/util/storage/models"
	b "bitbucket.org/level11consulting/ocelot/werker/builder"
	"bitbucket.org/level11consulting/ocelot/werker/valet"
	"fmt"
	"github.com/golang/protobuf/proto"
	"strings"

	//"runtime/debug"
	"context"
	"time"
	"bitbucket.org/level11consulting/ocelot/werker/config"
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

// watchForResults sends the *Transport object over the transport channel for stream functions to process
func (w *WorkerMsgHandler) WatchForResults(hash string, dbId int64) {
	ocelog.Log().Debugf("adding hash ( %s ) & infochan to transport channel", hash)
	transport := &Transport{Hash: hash, InfoChan: w.infochan, DbId: dbId}
	w.StreamChan <- transport
}


// MakeItSo will call appropriate builder functions
func (w *WorkerMsgHandler) MakeItSo(werk *pb.WerkerTask, builder b.Builder, finish, done chan int) {
	ocelog.Log().Debug("hash build ", werk.CheckoutHash)

	w.BuildValet.RegisterDoneChan(werk.CheckoutHash, done)
	defer w.BuildValet.MakeItSoDed(finish)
	defer w.BuildValet.UnregisterDoneChan(werk.CheckoutHash)
	defer func(){
		ocelog.Log().Info("calling done for nsqpb")
		done <- 1
	}()

	ctx, cancel := context.WithCancel(context.Background())

	//send build context off, build kills are performed by calling cancel on the cacellable context
	w.BuildCtxChan <- &BuildContext{
		Hash: werk.CheckoutHash,
		Context: ctx,
		CancelFunc: cancel,
	}

	defer cancel()
	defer func(){
		ocelog.Log().Info("closing infochan for ", werk.Id)
		close(w.infochan)
	}()

	w.WatchForResults(werk.CheckoutHash, werk.Id)

	//update consul with active build data
	consul := w.WerkConf.RemoteConfig.GetConsul()
	// if we can't register with consul, bail, just exit out. the maintainer will soon be pausing message flow anyway
	if err := w.BuildValet.StartBuild(consul, werk.CheckoutHash, werk.Id); err != nil {
		return
	}

	setupStart := time.Now()
	w.BuildValet.Reset("setup", werk.CheckoutHash)

	dockerIdChan := make (chan string)
	go w.listenForDockerUuid(dockerIdChan, werk.CheckoutHash)
	// do setup stage
	setupResult, dockerUUid := builder.Setup(ctx, w.infochan, dockerIdChan, werk, w.WerkConf.RemoteConfig, w.WerkConf.ServicePort)
	// this is the last defer, so it'll be the first thing to be run after the command is finished
	defer w.BuildValet.Cleaner.Cleanup(ctx, dockerUUid, w.infochan)
	setupDura := time.Now().Sub(setupStart)

	if err := storeStageToDb(w.Store, werk.Id, setupResult, setupStart, setupDura.Seconds()); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't store build output")
		return
	}
	if setupResult.Status == pb.StageResultVal_FAIL {
		handleFailure(setupResult, w.Store, "setup", setupDura, werk.Id)
		return
	}

	// do integration setup stage
	start := time.Now()
	integrationResult, dockerUUid := w.doIntegrations(ctx, werk, w.Store, builder, w.WerkConf.RemoteConfig, w.infochan)
	dura := time.Now().Sub(start)
	if err := storeStageToDb(w.Store, werk.Id, integrationResult, start, dura.Seconds()); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't store build output")
		return
	}
	if integrationResult.Status == pb.StageResultVal_FAIL {
		handleFailure(integrationResult, w.Store, "integration setup", dura, werk.Id)
		return
	}

	//all stages listed inside of the projects's ocelot.yml are executed + stored here
	fail := false
	start = time.Now()
	for _, stage := range werk.BuildConf.Stages {
		if shouldSkip, err := handleTriggers(werk.Branch, werk.Id, w.Store, stage); err != nil {
			return
		} else if shouldSkip {
			continue
		}
		w.BuildValet.Reset(stage.Name, werk.CheckoutHash)
		stageStart := time.Now()
		stageResult := builder.Execute(ctx, stage, w.infochan, werk.CheckoutHash)
		ocelog.Log().WithField("hash", werk.CheckoutHash).Info("finished stage: ", stage.Name)
		stageDura := time.Now().Sub(stageStart)

		if stageResult.Status == pb.StageResultVal_FAIL {
			fail = true
			if err := storeStageToDb(w.Store, werk.Id, stageResult, stageStart, stageDura.Seconds()); err != nil {
				ocelog.IncludeErrField(err).Error("couldn't store build output")
				return
			}
			break
		}


		if err := storeStageToDb(w.Store, werk.Id, stageResult, stageStart, stageDura.Seconds()); err != nil {
			ocelog.IncludeErrField(err).Error("couldn't store build output")
			return
		}

	}

	//update build_summary table
	dura = time.Now().Sub(start)
	if err := w.Store.UpdateSum(fail, dura.Seconds(), werk.Id); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't update summary in database")
		return
	}
	ocelog.Log().Debugf("finished building id %s", werk.CheckoutHash)
}

// handleTriggers deals with the triggers section of the of the stage. Right now we only support a list of branches that should "trigger" this stage to run.
//  if the trigger block exists, and current branch is in the list of trigger branches, the funciton will return a shouldSkip of false, signifying that the stage should execute.
//  If the current branch is not in the list of trigger branches, shouldSkip of true will be returned and the stage should not be executed.
//  If the stage is a shouldSkip, it will also save to the database that the stage will be skipped.
func handleTriggers(branch string, id int64, store storage.BuildStage, stage *pb.Stage) (shouldSkip bool, err error) {
	// null value of bool is false, so shouldSkip is false until told otherwise
	if stage.Trigger != nil {
		if len(stage.Trigger.Branches) == 0 {
			ocelog.Log().Info("fyi, got a trigger block with an empty list of branches. seems dumb.")
			// return false, the block is empty and there is nothing to check 
			return
		}
		if !branchOk(branch, stage.Trigger.Branches) {
			// not sure if we should store, but i think its good visibility especially for right now
			result := &pb.Result{stage.Name, pb.StageResultVal_PASS, "", []string{fmt.Sprintf("skipping stage because %s is not in the trigger branches list", branch)}}
			if err = storeStageToDb(store, id, result, time.Now(), 0); err != nil {
				ocelog.IncludeErrField(err).Error("couldn't store build output")
				return
			}
			// we could save to db, the branch running is not in the list of trigger branches, so we can flip the shouldSkip bool now.
			shouldSkip = true
			return
		}
		ocelog.Log().Debugf("building from trigger stage with branch %s. triggerBranches are %s", branch, strings.Join(stage.Trigger.Branches, ", "))
	}
	// will return false
	return
}

func (w *WorkerMsgHandler) listenForDockerUuid(dockerChan chan string, checkoutHash string) error {
	dockerUuid := <- dockerChan

	if err := buildruntime.RegisterBuild(w.WerkConf.RemoteConfig.GetConsul(), w.WerkConf.WerkerUuid.String(), checkoutHash, dockerUuid); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't register build")
		return err
	}

	return nil
}

//storeStageToDb is a helper function for storing stages to db - this runs on completion of every stage
func storeStageToDb(storage storage.BuildStage, buildId int64, stageResult *pb.Result, start time.Time, dur float64) error {
	err := storage.AddStageDetail(&models.StageResult{
		BuildId:       buildId,
		Stage:         stageResult.Stage,
		Status:        int(stageResult.Status),
		Error:         stageResult.Error,
		Messages:      stageResult.Messages,
		StartTime:     start,
		StageDuration: dur,
	})

	if err != nil {
		return err
	}

	return nil
}

func handleFailure(result *pb.Result, store storage.OcelotStorage, stageName string, duration time.Duration, id int64) {
	errStr := fmt.Sprintf("%s stage failed", stageName)
	if len(result.Error) > 0 {
		errStr = errStr + result.Error
	}
	ocelog.Log().Error(errStr)
	if err := store.UpdateSum(true, duration.Seconds(), id); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't update summary in database")
	}
}

// doIntegrations will run all the integrations that (one day) are pertinent to the task at hand.
func (w *WorkerMsgHandler) doIntegrations(ctx context.Context, werk *pb.WerkerTask, store storage.CredTable, bldr b.Builder, rc cred.CVRemoteConfig, logout chan[]byte) (result *pb.Result, id string) {
	accountName := strings.Split(werk.FullName, "/")[0]
	result = &pb.Result{}
	id = bldr.GetContainerId()
	var setupMessages []string
	stage := b.InitStageUtil("INTEGRATION_UTIL")
	result.Messages = setupMessages
	//only if the build tool is maven do we worry about settings.xml
	if werk.BuildConf.BuildTool == "maven" {
		result = bldr.IntegrationSetup(ctx, nexus.GetSettingsXml, bldr.WriteMavenSettingsXml, "maven", rc, accountName, stage, result.Messages, store, logout)
		if result.Status == pb.StageResultVal_FAIL {
			return
		}
	}

	result = bldr.IntegrationSetup(ctx, dockr.GetDockerConfig, bldr.WriteDockerJson, "docker login", rc, accountName, stage, result.Messages, store, logout)
	if result.Status == pb.StageResultVal_FAIL {
		return
	}
	result = bldr.IntegrationSetup(ctx, w.returnWerkerPort, bldr.DownloadKubectl, "kubectl download", rc, accountName, stage, result.Messages, store, logout)
	if result.Status == pb.StageResultVal_FAIL {
		return
	}
	result = bldr.IntegrationSetup(ctx, k8s.GetKubeConfig, bldr.InstallKubeconfig, "kubeconfig render", rc, accountName, stage, result.Messages, store, logout)
	if result.Status == pb.StageResultVal_FAIL {
		return
	}
	result.Messages = append(result.Messages, "completed integration util setup stage \u2713")
	return
}

func (w *WorkerMsgHandler) returnWerkerPort(rc cred.CVRemoteConfig, store storage.CredTable, accountName string) (string, error) {
	ocelog.Log().Debug("returning werker port")
	return  w.WerkConf.ServicePort, nil
}

func branchOk(branch string, buildBranches []string) bool {
	for _, goodBranch := range buildBranches {
		if goodBranch == branch {
			return true
		}
	}
	return false
}
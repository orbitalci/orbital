package launcher

import (
	"context"
	"fmt"
	"strings"
	"time"

	ocelog "bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/ocelot/build"
	"bitbucket.org/level11consulting/ocelot/build/integrations"
	"bitbucket.org/level11consulting/ocelot/build/integrations/sshkey"

	"bitbucket.org/level11consulting/ocelot/build/integrations/dockerconfig"
	"bitbucket.org/level11consulting/ocelot/build/integrations/kubeconf"
	"bitbucket.org/level11consulting/ocelot/build/integrations/nexusm2"
	"bitbucket.org/level11consulting/ocelot/build/valet"
	cred "bitbucket.org/level11consulting/ocelot/common/credentials"
	"bitbucket.org/level11consulting/ocelot/models"
	"bitbucket.org/level11consulting/ocelot/models/pb"
	"bitbucket.org/level11consulting/ocelot/storage"
)

// watchForResults sends the *Transport object over the transport channel for stream functions to process
func (w *launcher) WatchForResults(hash string, dbId int64) {
	ocelog.Log().Debugf("adding hash ( %s ) & infochan to transport channel", hash)
	transport := &models.Transport{Hash: hash, InfoChan: w.infochan, DbId: dbId}
	w.StreamChan <- transport
}


// MakeItSo will call appropriate builder functions
func (w *launcher) MakeItSo(werk *pb.WerkerTask, builder build.Builder, finish, done chan int) {
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
	w.BuildCtxChan <- &models.BuildContext{
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
	consul := w.RemoteConf.GetConsul()
	// if we can't register with consul, bail, just exit out. the maintainer will soon be pausing message flow anyway
	if err := w.BuildValet.StartBuild(consul, werk.CheckoutHash, werk.Id); err != nil {
		return
	}

	setupStart := time.Now()
	w.BuildValet.Reset("setup", werk.CheckoutHash)

	dockerIdChan := make (chan string)
	go w.listenForDockerUuid(dockerIdChan, werk.CheckoutHash)
	// do setup stage
	setupResult, dockerUUid := builder.Setup(ctx, w.infochan, dockerIdChan, werk, w.RemoteConf, w.ServicePort)
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
	integrationResult, dura, start := w.doIntegrations(ctx, werk, builder)
	if err := storeStageToDb(w.Store, werk.Id, integrationResult, start, dura.Seconds()); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't store build output")
		return
	}
	if integrationResult.Status == pb.StageResultVal_FAIL {
		handleFailure(integrationResult, w.Store, "integration setup", dura, werk.Id)
		return
	}
	// download necessary binaries (ie kubectl, w/e)
	downloadResult, duration, starttime := w.downloadBinaries(ctx, build.InitStageUtil("download binaries"), builder)
	if err := storeStageToDb(w.Store, werk.Id, downloadResult, starttime, duration.Seconds()); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't store build output")
		return
	}
	if downloadResult.Status == pb.StageResultVal_FAIL {
		handleFailure(downloadResult, w.Store, "download binaries", duration, werk.Id)
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

func (w *launcher) listenForDockerUuid(dockerChan chan string, checkoutHash string) error {
	dockerUuid := <- dockerChan

	if err := valet.RegisterBuild(w.RemoteConf.GetConsul(), w.Uuid.String(), checkoutHash, dockerUuid); err != nil {
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
func (w *launcher) doIntegrations(ctx context.Context, werk *pb.WerkerTask, bldr build.Builder) (result *pb.Result, duration time.Duration, start time.Time) {
	start = time.Now()
	defer func(){duration = time.Now().Sub(start)}()
	accountName := strings.Split(werk.FullName, "/")[0]
	result = &pb.Result{}
	var integMessages []string
	stage := build.InitStageUtil("INTEGRATION_UTIL")

	// todo: idk where to put this? where to instantiate integrations.. probably should just be a part of launcher?
	var integral = []integrations.StringIntegrator{sshkey.Create(), dockerconfig.Create(), kubeconf.Create(), nexusm2.Create()}
	for _, integ := range integral {
		if !integ.IsRelevant(werk.BuildConf) {
			continue
		}
		subStage := build.CreateSubstage(stage, integ.String())
		credz, err := w.RemoteConf.GetCredsBySubTypeAndAcct(w.Store, integ.SubType(), accountName, false)
		if err != nil {
			result = handleIntegrationErr(err, integ.String(), subStage, result.Messages)
			// if handleIntegrationError decides that this "failure" is actually OK, just continue to next integration
			if result.Status == pb.StageResultVal_PASS {
				integMessages = append(integMessages, result.Messages...)
				result.Messages = []string{}
				continue
			}
			return
		}
		integString, err := integ.GenerateIntegrationString(credz)
		if err != nil {
			result = &pb.Result{
				Stage:    subStage.GetStage(),
				Status:   pb.StageResultVal_FAIL,
				Error:    err.Error(),
				Messages: integMessages,
			}
			return
		}
		stg := &pb.Stage{Env: integ.GetEnv(), Script: integ.MakeBashable(integString), Name: subStage.Stage}
		result = bldr.ExecuteIntegration(ctx, stg, subStage, w.infochan)
		if result.Status == pb.StageResultVal_FAIL {
			result.Messages = append(integMessages, result.Messages...)
			return
		}
		integMessages = append(integMessages, result.Messages...)
		result.Messages = []string{}
		//integMessages = append(integMessages, "finished integration setup for " + subStage.Stage)
	}
	// reset stage to integration_util
	result.Stage = stage.GetStage()
	result.Messages = append(integMessages, "completed integration util setup stage \u2713")
	return
}

func (w *launcher) downloadBinaries(ctx context.Context, su *build.StageUtil, bldr build.Builder) (result *pb.Result, duration time.Duration, start time.Time) {
	start = time.Now()
	defer func(){duration = time.Now().Sub(start)}()
	// todo: there wil likely be more binaries to download in the future, should probably use the same pattern
	// as StringIntegrator.. maybe a DownloadIntegrator?
	subStage := build.CreateSubstage(su, "kubectl download")
	kubectl := &pb.Stage{Env: []string{}, Script: bldr.DownloadKubectl(w.ServicePort), Name: subStage.Stage,}
	result = bldr.ExecuteIntegration(ctx, kubectl, subStage, w.infochan)
	if result.Status == pb.StageResultVal_FAIL {
		return
	}
	result.Messages = append(result.Messages, "finished " + subStage.Stage)
	return
}

func (w *launcher) returnWerkerPort(rc cred.CVRemoteConfig, store storage.CredTable, accountName string) (string, error) {
	ocelog.Log().Debug("returning werker port")
	return  w.ServicePort, nil
}

func handleIntegrationErr(err error, integrationName string, stage *build.StageUtil, msgs []string) *pb.Result {
	_, ok := err.(*integrations.NoCreds)
	if !ok {
		ocelog.IncludeErrField(err).Error("returning failed setup because repo integration failed for: ", integrationName)
		return &pb.Result{
			Stage: stage.GetStage(),
			Status: pb.StageResultVal_FAIL,
			Error: err.Error(),
		}
	} else {
		msgs = append(msgs, "no integration data found for " + integrationName + " so assuming integration not necessary")
		return &pb.Result{
			Stage: stage.GetStage(),
			Status: pb.StageResultVal_PASS,
			Error: "",
			Messages: msgs,
		}
	}
}

func branchOk(branch string, buildBranches []string) bool {
	for _, goodBranch := range buildBranches {
		if goodBranch == branch {
			return true
		}
	}
	return false
}

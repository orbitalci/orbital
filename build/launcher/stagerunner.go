package launcher

import (
	"context"
	"fmt"
	"strings"
	"time"

	ocelog "github.com/shankj3/go-til/log"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"
)

func (w *launcher) runStages(ctx context.Context, werk *pb.WerkerTask, builder build.Builder) (fail bool, dura time.Duration, err error) {
	//all stages listed inside of the projects's ocelot.yml are executed + stored here
	fail = false
	start := time.Now()
	for _, stage := range werk.BuildConf.Stages {
		var shouldSkip bool
		if shouldSkip, err = handleTriggers(werk.Branch, werk.Id, w.Store, stage); err != nil {
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
			if err = storeStageToDb(w.Store, werk.Id, stageResult, stageStart, stageDura.Seconds()); err != nil {
				ocelog.IncludeErrField(err).Error("couldn't store build output")
				return
			}
			break
		}


		if err = storeStageToDb(w.Store, werk.Id, stageResult, stageStart, stageDura.Seconds()); err != nil {
			ocelog.IncludeErrField(err).Error("couldn't store build output")
			return
		}

	}

	dura = time.Now().Sub(start)
	return
}


// branchOk is an "if elem in list" check.
func branchOk(branch string, buildBranches []string) bool {
	for _, goodBranch := range buildBranches {
		if goodBranch == branch {
			return true
		}
	}
	return false
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
		branchGood, err := build.BranchRegexOk(branch, stage.Trigger.Branches)
		if err != nil {
			result := &pb.Result{Stage: stage.Name, Status: pb.StageResultVal_FAIL, Error: err.Error(), Messages: []string{"failed to check if current branch fit the trigger criteria"}}
			// not sure if we should store, but i think its good visibility especially for right now
			if err = storeStageToDb(store, id, result, time.Now(), 0); err != nil {
				ocelog.IncludeErrField(err).Error("couldn't store build output")
				return shouldSkip, err
			}
		}
		if !branchGood {
			// todo: have a SKIPPED value? more descriptive
			result := &pb.Result{Stage: stage.Name, Status: pb.StageResultVal_PASS, Error: "", Messages:[]string{fmt.Sprintf("skipping stage because %s is not in the trigger branches list", branch)}}
			// we could save to db, the branch running is not in the list of trigger branches, so we can flip the shouldSkip bool now.
			shouldSkip = true
			if err = storeStageToDb(store, id, result, time.Now(), 0); err != nil {
				ocelog.IncludeErrField(err).Error("couldn't store build output")
				return shouldSkip, err
			}
			return shouldSkip, err
		}
		ocelog.Log().Debugf("building from trigger stage with branch %s. triggerBranches are %s", branch, strings.Join(stage.Trigger.Branches, ", "))
	}
	// will return false
	return
}
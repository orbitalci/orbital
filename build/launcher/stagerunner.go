package launcher

import (
	"context"
	"strings"
	"time"

	ocelog "github.com/shankj3/go-til/log"
	"github.com/level11consulting/ocelot/build"
	"github.com/level11consulting/ocelot/common/trigger"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/storage"
)

// runStages runs the stages that are defined by the **user** ie what is in the ocelot.yml. Results will be stored in this stage as well.
//   runStages will return whether any of the stages failed, the amount of time it took to run all the stages, and errors (if any)
func (w *launcher) runStages(ctx context.Context, werk *pb.WerkerTask, builder build.Builder) (fail bool, dura time.Duration, err error) {
	//all stages listed inside of the projects's ocelot.yml are executed + stored here
	fail = false
	start := time.Now()
	for _, stage := range werk.BuildConf.Stages {
		var shouldSkip bool
		if shouldSkip, err = handleTriggers(werk, w.Store, stage); err != nil {
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

// handleTriggers deals with the triggers section of the of the stage. Right now we only support a list of branches that should "trigger" this stage to run.
//  if the trigger block exists, and current branch is in the list of trigger branches, the funciton will return a shouldSkip of false, signifying that the stage should execute.
//  If the current branch is not in the list of trigger branches, shouldSkip of true will be returned and the stage should not be executed.
//  If the stage is a shouldSkip, it will also save to the database that the stage will be skipped.
//  If the triggers block exists, and none of the conditions match the files changed / commit messages / branch, then the stage will be skipped
//  `triggers` is the new way. the old way, the `trigger` block in the yaml with the branch specification, will only run if the new `triggers` is nil
func handleTriggers(task *pb.WerkerTask, store storage.BuildStage, stage *pb.Stage) (shouldSkip bool, err error) {
	// null value of bool is false, so shouldSkip is false until told otherwise
	// only check for older version of trigger block if newer version, "triggers", is null because trigger is being deprecated
	if stage.Trigger != nil && stage.Triggers == nil {
		if len(stage.Trigger.Branches) == 0 {
			ocelog.Log().Info("fyi, got a trigger block with an empty list of branches. seems dumb.")
			// return false, the block is empty and there is nothing to check
			return
		}
		branchGood, err := trigger.BranchRegexOk(task.Branch, stage.Trigger.Branches)
		if err != nil {
			result := &pb.Result{Stage: stage.Name, Status: pb.StageResultVal_FAIL, Error: err.Error(), Messages: []string{"failed to check if current branch fit the trigger criteria"}}
			// not sure if we should store, but i think its good visibility especially for right now
			if err = storeStageToDb(store, task.Id, result, time.Now(), 0); err != nil {
				ocelog.IncludeErrField(err).Error("couldn't store build output")
				return shouldSkip, err
			}
		}
		// if no branches match and the trigger block was set, then skip the stage
		if !branchGood {
			return true, storeSkipped(store, stage, task.Id)
		}
		ocelog.Log().Debugf("building from trigger stage with branch %s. triggerBranches are %s", task.Branch, strings.Join(stage.Trigger.Branches, ", "))
	}
	// each line in triggers is an OR statement, so if any of them are fulfilled, then run the stage
	if stage.Triggers != nil {
		var passed bool
		for _, triggerString := range stage.Triggers {
			var condition *trigger.ConditionalDirective
			condition, err = trigger.Parse(triggerString)
			if err != nil {
				return
			}
			if condition.IsFulfilled(task.ChangesetData) {
				passed = true
				break
			}
		}
		// if none of the conditions in the triggers list were met, then skip the stage
		shouldSkip = !passed
		if shouldSkip {
			return true, storeSkipped(store, stage, task.Id)
		}
	}
	return
}


func storeSkipped(store storage.BuildStage, stage *pb.Stage, id int64) (err error) {
	result := &pb.Result{Stage: stage.Name, Status: pb.StageResultVal_SKIP, Error: "", Messages: []string{"skipping because the current changeset does not meet the trigger conditions"}}
	if err = storeStageToDb(store, id, result, time.Now(), 0); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't store build output")
		return
	}
	return
}

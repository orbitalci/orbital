package launcher

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	ocelog "github.com/shankj3/go-til/log"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/common"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
)

//preFlight will run all of the ocelot-specific setup stages that are not explicitly tied to the builder implementations' setup.
//   it will do:
//		- all the integrations (kubeconfig render, etc...)
//		- download binaries (kubectl, etc...)
//		- download codebase for building
//   all the substages above will be rolled into one stage for storage: PREFLIGHT
func (w *launcher) preFlight(ctx context.Context, werk *pb.WerkerTask, builder build.Builder) (bailOut bool, err error) {
	start := time.Now()
	prefly := build.InitStageUtil("PREFLIGHT")
	preflightResult := &pb.Result{Stage: prefly.GetStage(), Status: pb.StageResultVal_PASS}
	acct, _, err := common.GetAcctRepo(werk.FullName)
	if err != nil {
		return bailOut, err
	}
	var result *pb.Result
	result = w.handleEnvSecrets(ctx, builder, acct, prefly)
	if bailOut, err = w.mapOrStoreStageResults(result, preflightResult, werk.Id, start); err != nil || bailOut {
		return
	}
	result = w.downloadBinaries(ctx, prefly, builder, werk.BuildConf)
	if bailOut, err = w.mapOrStoreStageResults(result, preflightResult, werk.Id, start); err != nil || bailOut {
		return
	}
	result = w.doIntegrations(ctx, werk, builder, prefly)
	if bailOut, err = w.mapOrStoreStageResults(result, preflightResult, werk.Id, start); err != nil || bailOut {
		return
	}

	// download codebase to werker node
	result = downloadCodebase(ctx, werk, builder, prefly, w.infochan)
	if bailOut, err = w.mapOrStoreStageResults(result, preflightResult, werk.Id, start); err != nil || bailOut {
		return
	}
	if err = storeStageToDb(w.Store, werk.Id, preflightResult, start, time.Since(start).Seconds()); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't store build stage details for preflight")
		return
	}
	return
}

// mapOrStoreStageResults will map the subStageResult's messages to the parentResult. if the substageresult fails,
//   the newly mapped parentResult will be stored in OcelotStorage and bailOut will be returned as true.
//   if the storage fails, an error will be returned
func (w *launcher) mapOrStoreStageResults(subStageResult *pb.Result, parentResult *pb.Result, id int64, start time.Time) (bailOut bool, err error) {
	bailOut = subStageResult.Status == pb.StageResultVal_FAIL
	var preppedmessages []string
	for _, msg := range subStageResult.Messages {
		preppedmessages = append(preppedmessages, build.InitStageUtil(subStageResult.Stage).GetStageLabel()+msg)
	}
	parentResult.Messages = append(parentResult.Messages, subStageResult.Messages...)
	if bailOut {
		parentResult.Status = subStageResult.Status
		parentResult.Error = subStageResult.Error
		if err = storeStageToDb(w.Store, id, parentResult, start, time.Since(start).Seconds()); err != nil {
			ocelog.IncludeErrField(err).Errorf("failed to store for parent stage %s at substage %s", parentResult.Stage, subStageResult.Stage)
			return
		}
	}
	return
}

// handleEnvSecrets will grab all environment type credentials from storage / secret store and add them as global environment variables for
// 	the entire build.
func (w *launcher) handleEnvSecrets(ctx context.Context, builder build.Builder, accountName string, stage *build.StageUtil) *pb.Result {
	creds, err := w.RemoteConf.GetCredsBySubTypeAndAcct(w.Store, pb.SubCredType_ENV, accountName, false)
	if err != nil {
		if _, ok := err.(*common.NoCreds); ok {
			return &pb.Result{Status: pb.StageResultVal_PASS, Messages: []string{fmt.Sprintf("no env vars for %s %s", accountName, models.CHECKMARK)}, Stage: stage.GetStage()}
		}
		return &pb.Result{Error: err.Error(), Status: pb.StageResultVal_FAIL, Messages: []string{"could not get env secrets " + models.FAILED}, Stage: stage.GetStage()}
	}
	var allenvs []string
	for _, envVar := range creds {
		allenvs = append(allenvs, fmt.Sprintf("%s=%s", envVar.GetIdentifier(), envVar.GetClientSecret()))
	}
	builder.AddGlobalEnvs(allenvs)
	return &pb.Result{Status: pb.StageResultVal_PASS, Messages: []string{"successfully set env secrets " + models.CHECKMARK}, Stage: stage.GetStage()}
}

//downloadCodebase will download the code that will be built
func downloadCodebase(ctx context.Context, task *pb.WerkerTask, builder build.Builder, su *build.StageUtil, logChan chan []byte) *pb.Result {
	var setupMessages []string
	setupMessages = append(setupMessages, "attempting to download codebase...")
	stage := &pb.Stage{
		Env:    []string{},
		Script: builder.DownloadCodebase(task),
		Name:   su.GetStage(),
	}
	codebaseDownload := builder.ExecuteIntegration(ctx, stage, su, logChan)
	codebaseDownload.Messages = append(setupMessages, codebaseDownload.Messages...)
	if len(codebaseDownload.Error) > 0 {
		ocelog.Log().Error("an err happened trying to download codebase", codebaseDownload.Error)
	}
	return codebaseDownload
}

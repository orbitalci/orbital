package runtime

import (
	"context"
	"fmt"
	"time"

	builderinterface "github.com/level11consulting/orbitalci/build/builder/interface"
	accountrepo "github.com/level11consulting/orbitalci/build/helpers/stringbuilder/accountrepo"
	nocreds "github.com/level11consulting/orbitalci/build/helpers/stringbuilder/nocreds"
	"github.com/level11consulting/orbitalci/models"
	"github.com/level11consulting/orbitalci/models/pb"
	ocelog "github.com/shankj3/go-til/log"
)

//preFlight will run all of the ocelot-specific setup stages that are not explicitly tied to the builder implementations' setup.
//   it will do:
//		- all the integrations (kubeconfig render, etc...)
//		- download binaries (kubectl, etc...)
//		- download codebase for building
//   all the substages above will be rolled into one stage for storage: PREFLIGHT
func (w *launcher) preFlight(ctx context.Context, werk *pb.WerkerTask, builder builderinterface.Builder) (bailOut bool, err error) {
	ocelog.Log().Debug("In Preflight stage")
	start := time.Now()
	prefly := builderinterface.InitStageUtil("PREFLIGHT")
	preflightResult := &pb.Result{Stage: prefly.GetStage(), Status: pb.StageResultVal_PASS}
	acct, _, err := accountrepo.GetAcctRepo(werk.FullName)
	if err != nil {
		ocelog.IncludeErrField(err).Error("Failed GetAcctRepo")
		return bailOut, err
	}
	var result *pb.Result
	ocelog.Log().Debug("Getting env secrets")
	result = w.handleEnvSecrets(ctx, builder, acct, prefly)
	if bailOut, err = w.mapOrStoreStageResults(result, preflightResult, werk.Id, start); err != nil || bailOut {
		ocelog.IncludeErrField(err).Error("Failed mapOrStoreStageResults")
		return
	}
	ocelog.Log().Debug("Downloading binaries")
	result = w.downloadBinaries(ctx, prefly, builder, werk.BuildConf)
	if bailOut, err = w.mapOrStoreStageResults(result, preflightResult, werk.Id, start); err != nil || bailOut {
		ocelog.IncludeErrField(err).Error("Failed mapOrStoreStageResults")
		return
	}
	ocelog.Log().Debug("Doing integrations")
	result = w.doIntegrations(ctx, werk, builder, prefly)
	if bailOut, err = w.mapOrStoreStageResults(result, preflightResult, werk.Id, start); err != nil || bailOut {
		ocelog.IncludeErrField(err).Error("Failed mapOrStoreStageResults")
		return
	}

	// download codebase to werker node
	ocelog.Log().Debug("Downloading codebase")
	result = downloadCodebase(ctx, werk, builder, prefly, w.infochan)
	if bailOut, err = w.mapOrStoreStageResults(result, preflightResult, werk.Id, start); err != nil || bailOut {
		ocelog.IncludeErrField(err).Error("Failed mapOrStoreStageResults")
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
		preppedmessages = append(preppedmessages, builderinterface.InitStageUtil(subStageResult.Stage).GetStageLabel()+msg)
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
func (w *launcher) handleEnvSecrets(ctx context.Context, builder builderinterface.Builder, accountName string, stage *builderinterface.StageUtil) *pb.Result {

	ocelog.Log().Debug("Getting env vars")

	return &pb.Result{Status: pb.StageResultVal_PASS, Messages: []string{"This is when environment vars get set in the build environment -- skipping... " + models.CHECKMARK}, Stage: stage.GetStage()}

	creds, err := w.RemoteConf.GetCredsBySubTypeAndAcct(w.Store, pb.SubCredType_ENV, accountName, false)
	if err != nil {
		if _, ok := err.(*nocreds.NoCreds); ok {
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
func downloadCodebase(ctx context.Context, task *pb.WerkerTask, builder builderinterface.Builder, su *builderinterface.StageUtil, logChan chan []byte) *pb.Result {
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

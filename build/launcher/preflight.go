package launcher

import (
	"context"
	"errors"
	"time"

	ocelog "bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/ocelot/build"
	"bitbucket.org/level11consulting/ocelot/models/pb"
)

//preFlight will run all of the ocelot-specific setup stages that are not explicitly tied to the builder implementations' setup.
//   it will do:
//		- all the integrations (kubeconfig render, etc...)
//		- download binaries (kubectl, etc...)
//		- download codebase for building
func (w *launcher) preFlight(ctx context.Context, werk *pb.WerkerTask, builder build.Builder) (err error){
	// do integration setup stage
	integrationResult, dura, start := w.doIntegrations(ctx, werk, builder)
	if err = storeStageToDb(w.Store, werk.Id, integrationResult, start, dura.Seconds()); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't store build output")
		return
	}
	if integrationResult.Status == pb.StageResultVal_FAIL {
		handleFailure(integrationResult, w.Store, "integration setup", dura, werk.Id)
		err = errors.New("integration stage failed")
		return
	}

	//// download necessary binaries (ie kubectl, w/e)
	//downloadResult, duration, starttime := w.downloadBinaries(ctx, build.InitStageUtil("download binaries"), builder)
	//if err = storeStageToDb(w.Store, werk.Id, downloadResult, starttime, duration.Seconds()); err != nil {
	//	ocelog.IncludeErrField(err).Error("couldn't store build output")
	//	return
	//}
	//if downloadResult.Status == pb.StageResultVal_FAIL {
	//	handleFailure(downloadResult, w.Store, "download binaries", duration, werk.Id)
	//	err = errors.New("extra binaries download failed")
	//	return
	//}

	// download codebase to werker node
	codeResult, duration, starttime := downloadCodebase(ctx, werk, builder, w.infochan)
	if err = storeStageToDb(w.Store, werk.Id, codeResult, starttime, duration.Seconds()); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't store build stage details")
		return
	}
	if codeResult.Status == pb.StageResultVal_FAIL {
		handleFailure(codeResult, w.Store, "download codebase", duration, werk.Id)
		err = errors.New("failed to download codebase for build")
		return
	}
	return
}

//downloadCodebase will download the code that will be built
func downloadCodebase(ctx context.Context, task *pb.WerkerTask, builder build.Builder, logChan chan []byte) (*pb.Result, time.Duration, time.Time) {
	start := time.Now()
	var setupMessages []string
	setupMessages = append(setupMessages, "attempting to download codebase...")
	su := build.InitStageUtil("code download")
	stage := &pb.Stage{
		Env: []string{},
		Script: builder.DownloadCodebase(task),
		Name: su.GetStage(),
	}
	codebaseDownload := builder.ExecuteIntegration(ctx, stage, su, logChan)
	codebaseDownload.Messages = append(setupMessages, codebaseDownload.Messages...)
	if len(codebaseDownload.Error) > 0 {
		ocelog.Log().Error("an err happened trying to download codebase", codebaseDownload.Error)
	}
	return codebaseDownload, time.Now().Sub(start), start
}

package launcher

import (
	"context"
	"strings"
	"time"

	"github.com/shankj3/go-til/log"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/build/integrations"
	"github.com/shankj3/ocelot/build/integrations/dockerconfig"
	"github.com/shankj3/ocelot/build/integrations/kubeconf"
	"github.com/shankj3/ocelot/build/integrations/nexusm2"
	"github.com/shankj3/ocelot/build/integrations/sshkey"
	"github.com/shankj3/ocelot/build/integrations/xcode"
	"github.com/shankj3/ocelot/common"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
)

// todo: idk where to put this? where to instantiate integrations.. probably should just be a part of launcher?
func getIntegrationList() []integrations.StringIntegrator {
	return []integrations.StringIntegrator{
		sshkey.Create(),
		dockerconfig.Create(),
		kubeconf.Create(),
		nexusm2.Create(),
		xcode.Create(),
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

	var integral = getIntegrationList()
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
	result.Messages = append(integMessages, "completed integration util setup stage " + models.CHECKMARK)
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
		result.Messages = append(result.Messages, "failed to download kubectl, continuing anyway as it may not be used...")
		result.Status = pb.StageResultVal_PASS
	}
	result.Messages = append(result.Messages, "finished " + subStage.Stage)
	return
}


func handleIntegrationErr(err error, integrationName string, stage *build.StageUtil, msgs []string) *pb.Result {
	_, ok := err.(*common.NoCreds)
	if !ok {
		log.IncludeErrField(err).Error("returning failed setup because repo integration failed for: ", integrationName)
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

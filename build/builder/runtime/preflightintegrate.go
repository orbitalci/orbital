package runtime

import (
	"context"
	"fmt"

	builderinterface "github.com/level11consulting/orbitalci/build/builder/interface"
	accountrepo "github.com/level11consulting/orbitalci/build/helpers/stringbuilder/accountrepo"
	nocreds "github.com/level11consulting/orbitalci/build/helpers/stringbuilder/nocreds"
	"github.com/level11consulting/orbitalci/build/integrations"
	"github.com/level11consulting/orbitalci/build/integrations/dockerconfig"
	"github.com/level11consulting/orbitalci/build/integrations/helm"
	"github.com/level11consulting/orbitalci/build/integrations/helmrepo"
	"github.com/level11consulting/orbitalci/build/integrations/kubeconf"
	"github.com/level11consulting/orbitalci/build/integrations/kubectl"
	"github.com/level11consulting/orbitalci/build/integrations/minio"
	"github.com/level11consulting/orbitalci/build/integrations/minioconfig"
	"github.com/level11consulting/orbitalci/build/integrations/nexusm2"
	"github.com/level11consulting/orbitalci/build/integrations/sshkey"
	"github.com/level11consulting/orbitalci/build/integrations/xcode"
	"github.com/level11consulting/orbitalci/models"
	"github.com/level11consulting/orbitalci/models/pb"
	ocelog "github.com/shankj3/go-til/log"
)

// todo: idk where to put this? where to instantiate integrations.. probably should just be a part of launcher?
func getIntegrationList() []integrations.StringIntegrator {
	return []integrations.StringIntegrator{
		sshkey.Create(),
		dockerconfig.Create(),
		kubeconf.Create(),
		nexusm2.Create(),
		xcode.Create(),
		minioconfig.Create(),
		helmrepo.Create(),
	}
}

func getBinaryIntegList(loopbackHost, loopbackPort string) []integrations.BinaryIntegrator {
	return []integrations.BinaryIntegrator{
		kubectl.Create(loopbackHost, loopbackPort),
		helm.Create(loopbackHost, loopbackPort),
		minio.Create(loopbackHost, loopbackPort),
	}
}

// doIntegrations will run all the integrations that (one day) are pertinent to the task at hand.
func (w *launcher) doIntegrations(ctx context.Context, werk *pb.WerkerTask, bldr builderinterface.Builder, baseStage *builderinterface.StageUtil) (result *pb.Result) {
	accountName, _, err := accountrepo.GetAcctRepo(werk.FullName)
	if err != nil {
		result.Status = pb.StageResultVal_FAIL
		result.Error = err.Error()
		return
	}
	result = &pb.Result{}
	var integMessages []string
	stage := builderinterface.CreateSubstage(baseStage, "INTEG")
	for _, integ := range w.integrations {
		if !integ.IsRelevant(werk.BuildConf) {
			continue
		}
		subStage := builderinterface.CreateSubstage(stage, integ.String())
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
	result.Messages = append(integMessages, "completed integration util setup stage "+models.CHECKMARK)
	return
}

// downloadBinaries runs all the binary integrations that are associated with this build. the binary integrations are defined by
//  getBinaryIntegList.
func (w *launcher) downloadBinaries(ctx context.Context, su *builderinterface.StageUtil, bldr builderinterface.Builder, wc *pb.BuildConfig) (result *pb.Result) {
	var integMessages []string
	result = &pb.Result{}
	for _, binaryI := range w.binaryIntegs {
		subStage := builderinterface.CreateSubstage(su, binaryI.String())
		if binaryI.IsRelevant(wc) {
			stg := &pb.Stage{Name: subStage.Stage, Script: binaryI.GenerateDownloadBashables()}
			result = bldr.ExecuteIntegration(ctx, stg, subStage, w.infochan)
			integMessages = append(integMessages, result.Messages...)
			if result.Status == pb.StageResultVal_FAIL {
				result.Messages = integMessages
				return
			}
			result.Messages = []string{}
		} else {
			integMessages = append(integMessages, fmt.Sprintf("%s not relevant", binaryI.String()))
		}
	}
	// reset stage to download_binaries
	result.Stage = su.GetStage()
	result.Messages = append(integMessages, "completed download binaries setup stage "+models.CHECKMARK)
	return
}

func handleIntegrationErr(err error, integrationName string, stage *builderinterface.StageUtil, msgs []string) *pb.Result {
	_, ok := err.(*nocreds.NoCreds)
	if !ok {
		ocelog.IncludeErrField(err).Error("returning failed setup because repo integration failed for: ", integrationName)
		return &pb.Result{
			Stage:    stage.GetStage(),
			Status:   pb.StageResultVal_FAIL,
			Error:    err.Error(),
			Messages: append(msgs, fmt.Sprintf("integration failed for %s %s", integrationName, models.FAILED)),
		}
	} else {
		msgs = append(msgs, fmt.Sprintf("no integration data for %s %s", integrationName, models.CHECKMARK))
		return &pb.Result{
			Stage:    stage.GetStage(),
			Status:   pb.StageResultVal_PASS,
			Error:    "",
			Messages: msgs,
		}
	}
}

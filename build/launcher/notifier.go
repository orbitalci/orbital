package launcher

import (
	"context"
	"strings"
	"time"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/build/notifiers"
	"github.com/shankj3/ocelot/build/notifiers/slack"

)

func getNotifiers() []notifiers.Notifier {
	return []notifiers.Notifier{slack.Create()}
}

// doIntegrations will run all the integrations that (one day) are pertinent to the task at hand.
func (w *launcher) doNotifications(ctx context.Context, werk *pb.WerkerTask) error {
	accountName := strings.Split(werk.FullName, "/")[0]
	result = &pb.Result{}
	var integMessages []string

	// todo: idk where to put this? where to instantiate integrations.. probably should just be a part of launcher?
	//var integral = []integrations.StringIntegrator{sshkey.Create(), dockerconfig.Create(), kubeconf.Create(), nexusm2.Create()}
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
	result.Messages = append(integMessages, "completed integration util setup stage "+models.CHECKMARK)
	return
}

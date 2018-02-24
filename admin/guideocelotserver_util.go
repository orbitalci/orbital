package admin

import (
	"bitbucket.org/level11consulting/ocelot/util/handler"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	"github.com/golang/protobuf/ptypes/timestamp"
	adminModel "bitbucket.org/level11consulting/ocelot/admin/models"
	storeModel "bitbucket.org/level11consulting/ocelot/util/storage/models"
)

//when new configurations are added to the config channel, create bitbucket client and webhooks
func SetupCredentials(gosss adminModel.GuideOcelotServer, config *adminModel.VCSCreds) error {
	gos := gosss.(*guideOcelotServer)
	//hehe right now we only have bitbucket
	switch config.Type {
	case "bitbucket":
		bitbucketClient := &ocenet.OAuthClient{}
		bitbucketClient.Setup(config)

		bbHandler := handler.GetBitbucketHandler(config, bitbucketClient)
		err := bbHandler.Walk()
		if err != nil {
			return err
		}
	}
	configPath := config.BuildCredPath(config.Type, config.AcctName)
	err := gos.RemoteConfig.AddCreds(configPath, config)
	return err
}

func SetupRepoCredentials(gosss adminModel.GuideOcelotServer, config *adminModel.RepoCreds) error {
	// todo: probably should do some kind of test f they are valid or not? is there a way to test these creds
	gos := gosss.(*guideOcelotServer)
	configPath := config.BuildCredPath(config.Type, config.AcctName)
	err := gos.RemoteConfig.AddCreds(configPath, config)
	return err
}

func ParseStagesByBuildId(buildSum storeModel.BuildSummary, stageResults []storeModel.StageResult) *adminModel.Status {
	var parsedStages []*adminModel.Stage
	for _, result := range stageResults {
		stageDupe := &adminModel.Stage{
			Stage: result.Stage,
			Error: result.Error,
			Status: int32(result.Status),
			Messages: result.Messages,
			StartTime: &timestamp.Timestamp{Seconds: result.StartTime.UTC().Unix()},
			StageDuration: result.StageDuration,
		}
		parsedStages = append(parsedStages, stageDupe)
	}

	hashStatus := &adminModel.Status{
		BuildSum: &adminModel.BuildSummary{
			Hash: buildSum.Hash,
			Failed: buildSum.Failed,
			BuildTime: &timestamp.Timestamp{Seconds: buildSum.BuildTime.UTC().Unix()},
			Account: buildSum.Account,
			BuildDuration: buildSum.BuildDuration,
			Repo: buildSum.Repo,
			Branch: buildSum.Branch,
			BuildId: buildSum.BuildId,
		},
		Stages: parsedStages,
	}

	return hashStatus
}
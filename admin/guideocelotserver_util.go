package admin

import (
	ocenet "bitbucket.org/level11consulting/go-til/net"
	adminModel "bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/util/handler"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	storeModel "bitbucket.org/level11consulting/ocelot/util/storage/models"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		go bbHandler.Walk() //spawning walk in a different thread because we don't want client to wait if there's a lot of repos/files to check
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

//ParseStagesByBuildId will combine the buildsummary + stages to a single object called "Status"
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


// handleStorageError  will attempt to decipher if err is not found. if so, iwll set the appropriate grpc status code and return new grpc status error
func handleStorageError(err error) error {
	if _, ok := err.(*storage.ErrNotFound); ok {
		return status.Error(codes.NotFound, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}
package admin

import (
	ocenet "bitbucket.org/level11consulting/go-til/net"
	adminModel "bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/handler"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	storeModel "bitbucket.org/level11consulting/ocelot/util/storage/models"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"errors"
)

//when new configurations are added to the config channel, create bitbucket client and webhooks
func SetupCredentials(gosss adminModel.GuideOcelotServer, config *adminModel.VCSCreds) error {
	gos := gosss.(*guideOcelotServer)
	//hehe right now we only have bitbucket
	switch config.SubType {
	case adminModel.SubCredType_BITBUCKET:
		bitbucketClient := &ocenet.OAuthClient{}
		bitbucketClient.Setup(config)

		bbHandler := handler.GetBitbucketHandler(config, bitbucketClient)
		go bbHandler.Walk() //spawning walk in a different thread because we don't want client to wait if there's a lot of repos/files to check
	default:
		return errors.New("currently only bitbucket is supported")
	}

	config.BuildAndSetIdentifier()
	err := gos.RemoteConfig.AddCreds(config)
	return err
}

func SetupRCCCredentials(remoteConf cred.CVRemoteConfig, config adminModel.OcyCredder) error {
	err := remoteConf.AddCreds(config)
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
			QueueTime: &timestamp.Timestamp{Seconds: buildSum.QueueTime.UTC().Unix()},
		},
		Stages: parsedStages,
	}

	return hashStatus
}

//RespWrap will wrap streaming messages in a LineResponse object to be sent by the server stream
func RespWrap(msg string) *adminModel.LineResponse {
	return &adminModel.LineResponse{OutputLine: msg}
}

// handleStorageError  will attempt to decipher if err is not found. if so, iwll set the appropriate grpc status code and return new grpc status error
func handleStorageError(err error) error {
	if _, ok := err.(*storage.ErrNotFound); ok {
		return status.Error(codes.NotFound, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}
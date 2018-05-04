package admin

import (
	ocenet "github.com/shankj3/go-til/net"

	cred "github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"

	"github.com/golang/protobuf/ptypes/timestamp"
	bb "github.com/shankj3/ocelot/common/remote/bitbucket"
	"github.com/shankj3/ocelot/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"errors"
)

var unsupported = errors.New("currently only bitbucket is supported")

//when new configurations are added to the config channel, create bitbucket client and webhooks
func SetupCredentials(gosss pb.GuideOcelotServer, config *pb.VCSCreds) error {
	gos := gosss.(*guideOcelotServer)
	//hehe right now we only have bitbucket
	switch config.SubType {
	case pb.SubCredType_BITBUCKET:
		bitbucketClient := &ocenet.OAuthClient{}
		bitbucketClient.Setup(config)

		bbHandler := bb.GetBitbucketHandler(config, bitbucketClient)
		go bbHandler.Walk() //spawning walk in a different thread because we don't want client to wait if there's a lot of repos/files to check
	default:
		return unsupported
	}

	config.Identifier = config.BuildIdentifier()
	//right now, we will always overwrite
	err := gos.RemoteConfig.AddCreds(gos.Storage, config, true)
	return err
}

func SetupRCCCredentials(remoteConf cred.CVRemoteConfig, store storage.CredTable, config pb.OcyCredder) error {
	//right now, we will always overwrite
	err := remoteConf.AddCreds(store, config, true)
	return err
}

//ParseStagesByBuildId will combine the buildsummary + stages to a single object called "Status"
func ParseStagesByBuildId(buildSum models.BuildSummary, stageResults []models.StageResult) *pb.Status {
	var parsedStages []*pb.StageStatus
	for _, result := range stageResults {
		stageDupe := &pb.StageStatus{
			StageStatus:   result.Stage,
			Error:         result.Error,
			Status:        int32(result.Status),
			Messages:      result.Messages,
			StartTime:     &timestamp.Timestamp{Seconds: result.StartTime.UTC().Unix()},
			StageDuration: result.StageDuration,
		}
		parsedStages = append(parsedStages, stageDupe)
	}

	hashStatus := &pb.Status{
		BuildSum: &pb.BuildSummary{
			Hash:          buildSum.Hash,
			Failed:        buildSum.Failed,
			BuildTime:     &timestamp.Timestamp{Seconds: buildSum.BuildTime.UTC().Unix()},
			Account:       buildSum.Account,
			BuildDuration: buildSum.BuildDuration,
			Repo:          buildSum.Repo,
			Branch:        buildSum.Branch,
			BuildId:       buildSum.BuildId,
			QueueTime:     &timestamp.Timestamp{Seconds: buildSum.QueueTime.UTC().Unix()},
		},
		Stages: parsedStages,
	}

	return hashStatus
}

//RespWrap will wrap streaming messages in a LineResponse object to be sent by the server stream
func RespWrap(msg string) *pb.LineResponse {
	return &pb.LineResponse{OutputLine: msg}
}

// handleStorageError  will attempt to decipher if err is not found. if so, iwll set the appropriate grpc status code and return new grpc status error
func handleStorageError(err error) error {
	if _, ok := err.(*storage.ErrNotFound); ok {
		return status.Error(codes.NotFound, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}

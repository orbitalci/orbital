package valet

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shankj3/go-til/consul"
	ocelog "github.com/shankj3/go-til/log"
	"github.com/shankj3/ocelot/common"
)

// Register will add all the appropriate build details that the admin needs to contact the werker for stream info
// will add:
// werkerLocation  = "ci/werker_location/%s" // %s is werker id
// ci/werker_location/<werkid> + werker_ip        = ip
// 		'' 			           + werker_grpc_port = grpcPort
// 		''				       + werker_ws_port   = wsPort
// returns a generated uuid for the werker
func Register(consulete *consul.Consulet, ip string, grpcPort string, wsPort string) (werkerId uuid.UUID, err error) {
	werkerId = uuid.New()
	strId := werkerId.String()
	if err = consulete.AddKeyValue(common.MakeWerkerIpPath(strId), []byte(ip)); err != nil {
		return
	}
	if err = consulete.AddKeyValue(common.MakeWerkerGrpcPath(strId), []byte(grpcPort)); err != nil {
		return
	}
	if err = consulete.AddKeyValue(common.MakeWerkerWsPath(strId), []byte(wsPort)); err != nil {
		return
	}
	return
}

func UnRegister(consulete *consul.Consulet, werkerId string) error {
	err := consulete.RemoveValues(common.MakeWerkerLocPath(werkerId))
	return err
}

func RegisterStartedBuild(consulete *consul.Consulet, werkerId string, gitHash string) error {
	if err := consulete.AddKeyValue(common.MakeBuildMapPath(gitHash), []byte(werkerId)); err != nil {
		return err
	}
	return nil
}

func RegisterBuild(consulete *consul.Consulet, werkerId string, gitHash string, dockerUuid string) error {
	ocelog.Log().WithField("werker_id", werkerId).WithField("git_hash", gitHash).WithField("docker_uuid", dockerUuid).Info("registering build")
	err := consulete.AddKeyValue(common.MakeDockerUuidPath(werkerId, gitHash), []byte(dockerUuid))
	return err
}

func RegisterBuildSummaryId(consulete *consul.Consulet, werkerId string, gitHash string, buildId int64) error {
	str := fmt.Sprintf("%d", buildId)
	ocelog.Log().WithField("werker_id", werkerId).WithField("git_hash", gitHash).WithField("buildId", buildId).Info("registering build")
	err := consulete.AddKeyValue(common.MakeBuildSummaryIdPath(werkerId, gitHash), []byte(str))
	return err
}

func RegisterBuildStage(consulete *consul.Consulet, werkerId string, gitHash string, buildStage string) error {
	ocelog.Log().WithField("werker_id", werkerId).WithField("git_hash", gitHash).WithField("buildStage", buildStage).Info("registering build")
	err := consulete.AddKeyValue(common.MakeBuildStagePath(werkerId, gitHash), []byte(buildStage))
	return err
}

func RegisterStageStartTime(consulete *consul.Consulet, werkerId string, gitHash string, start time.Time) error {
	str := fmt.Sprintf("%d", start.Unix()) // todo: figure out a better way to do this conversion using bit shifting or something because i know this isnt the "right" way to do it
	err := consulete.AddKeyValue(common.MakeBuildStartpath(werkerId, gitHash), []byte(str))
	return err
}

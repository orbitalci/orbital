package build

import (
	"context"
	"strings"

	"bitbucket.org/level11consulting/ocelot/build/basher"
	cred "bitbucket.org/level11consulting/ocelot/common/credentials"
	pb "bitbucket.org/level11consulting/ocelot/models/pb"
	"bitbucket.org/level11consulting/ocelot/storage"
)

type RepoSetupFunc func(rc cred.CVRemoteConfig, store storage.CredTable, accountName string) (string, error)
type RepoExecFunc func(string) []string

type Builder interface {
	basher.OcyBash
	Setup(ctx context.Context, logout chan []byte, dockerId chan string, werk *pb.WerkerTask, rc cred.CVRemoteConfig, werkerPort string) (res *pb.Result, uuid string)
	Execute(ctx context.Context, actions *pb.Stage, logout chan []byte, commitHash string) *pb.Result
	GetContainerId() string
	IntegrationSetup(ctx context.Context, setupFunc RepoSetupFunc, execFunc RepoExecFunc, integrationName string, rc cred.CVRemoteConfig, accountName string, su *StageUtil, msgs []string, store storage.CredTable, logout chan []byte) (result *pb.Result)
}

//helper functions for stages, doesn't handle camelcase right now so if you want that set the values
//yourself explicitly
func InitStageUtil(stage string) *StageUtil {
	su := &StageUtil{
		Stage: strings.ToLower(stage),
		StageLabel: strings.ToUpper(stage) + " | ",
	}
	return su
}

type StageUtil struct {
	Stage string
	StageLabel string
}

func (s *StageUtil) GetStage() string {
	return s.Stage
}

func (s *StageUtil) GetStageLabel() string {
	return s.StageLabel
}

func (s *StageUtil) SetStage(stage string) {
	s.Stage = stage
}

func (s *StageUtil) SetStageLabel(stageLabel string) {
	s.StageLabel = stageLabel
}

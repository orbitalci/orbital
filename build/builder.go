package build

import (
	"context"
	"fmt"
	"io"
	"strings"

	pb "github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/server/config"
	"github.com/level11consulting/ocelot/storage"
)

type RepoSetupFunc func(rc config.CVRemoteConfig, store storage.CredTable, accountName string) (string, error)
type RepoExecFunc func(string) []string

type Builder interface {
	OcyBash
	Init(ctx context.Context, hash string, logout chan []byte) *pb.Result
	// for during setup stage, on instantiation of build
	SetGlobalEnv(envs []string)
	// for after that, if an integration calls for an environment variable to be set (i.e. the creds option for uploading env vars for use)
	AddGlobalEnvs(envs []string)
	Setup(ctx context.Context, logout chan []byte, dockerId chan string, werk *pb.WerkerTask, rc config.CVRemoteConfig, werkerPort string) (res *pb.Result, uuid string)
	Execute(ctx context.Context, actions *pb.Stage, logout chan []byte, commitHash string) *pb.Result
	ExecuteIntegration(ctx context.Context, stage *pb.Stage, stgUtil *StageUtil, logout chan []byte) *pb.Result
	GetContainerId() string

	io.Closer
}

//helper functions for stages, doesn't handle camelcase right now so if you want that set the values
//yourself explicitly
func InitStageUtil(stage string) *StageUtil {
	su := &StageUtil{
		Stage:      strings.ToLower(stage),
		StageLabel: strings.ToUpper(stage) + " | ",
	}
	return su
}

type StageUtil struct {
	Stage      string
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

func CreateSubstage(stg *StageUtil, subStage string) *StageUtil {
	return &StageUtil{
		Stage:      fmt.Sprintf("%s | %s", stg.GetStage(), strings.ToLower(subStage)),
		StageLabel: fmt.Sprintf("%s | %s | ", strings.ToUpper(stg.GetStage()), strings.ToUpper(subStage)),
	}
}

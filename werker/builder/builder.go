package builder

import (
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"strings"
	"context"
)


type Builder interface {
	Setup(ctx context.Context, logout chan []byte, dockerId chan string, werk *pb.WerkerTask, rc cred.CVRemoteConfig, werkerPort string) (res *pb.Result, uuid string)
	Execute(ctx context.Context, actions *pb.Stage, logout chan []byte, commitHash string) *pb.Result
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

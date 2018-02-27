package builder

import (
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"strings"
)


type Builder interface {
	Setup(logout chan []byte, werk *pb.WerkerTask, rc cred.CVRemoteConfig, werkerPort string) (res *Result, uuid string)
	Execute(actions *pb.Stage, logout chan []byte, commitHash string) *Result
	Cleanup(logout chan []byte)
}

type Result struct {
	Stage    string
	Status   StageResultVal
	Error    error
	Messages []string
}

type StageResultVal int32

const (
	PASS	StageResultVal = 0
	FAIL	StageResultVal = 1
)

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

package builder

import (
	pb "bitbucket.org/level11consulting/ocelot/protos"
)


type Builder interface {
	Setup(logout chan []byte, werk *pb.WerkerTask) *Result
	Build(logout chan []byte, stage *pb.Stage, commitHash string) *Result
	SaveArtifact(logout chan []byte, stage *pb.WerkerTask, commitHash string) *Result
	Execute(stage string, actions *pb.Stage, logout chan []byte) *Result
	Cleanup()
}

type Result struct {
	Stage string
	Status StageResult
	Error  error
}

type StageResult int32

const (
	PASS	StageResult = 0
	FAIL	StageResult = 1
)
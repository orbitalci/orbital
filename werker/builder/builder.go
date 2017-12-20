package builder

import (
	pb "bitbucket.org/level11consulting/ocelot/protos"
)


//TODO: think about how deployment to nexus fits in
type Builder interface {
	Setup(logout chan []byte, werk *pb.WerkerTask) *Result
	Build(logout chan []byte, stage *pb.Stage, commitHash string) *Result
	Execute(stage string, actions *pb.Stage, logout chan []byte) *Result
	Cleanup()
}

//TODO: could return even less
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
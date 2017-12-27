package builder

import (
	pb "bitbucket.org/level11consulting/ocelot/protos"
)


//TODO: think about where deployment to nexus fits in - do we want this in its own method? Or will it be up to builder implementations to embed into their 'build' stage?
type Builder interface {
	Setup(logout chan []byte, werk *pb.WerkerTask) *Result
	Build(logout chan []byte, stage *pb.Stage, commitHash string) *Result
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
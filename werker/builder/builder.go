package builder

import (
	pb "bitbucket.org/level11consulting/ocelot/protos"
)


//TODO: think about where deployment to nexus fits in - do we want this in its own method? Or will it be up to builder implementations to embed into their 'build' stage?
type Builder interface {
	Setup(logout chan []byte, werk *pb.WerkerTask) *Result
	Execute(actions *pb.Stage, logout chan []byte, commitHash string) *Result
	Cleanup()
}

type Result struct {
	Stage    string
	Status   StageResult
	Error    error
	Messages []string
}

type StageResult int32

const (
	PASS	StageResult = 0
	FAIL	StageResult = 1
)
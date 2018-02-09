package builder

import (
	pb "bitbucket.org/level11consulting/ocelot/protos"
)


type Builder interface {
	Setup(logout chan []byte, werk *pb.WerkerTask) *Result
	Execute(actions *pb.Stage, logout chan []byte, commitHash string) *Result
	Cleanup()
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
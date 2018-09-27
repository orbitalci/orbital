package build_signaler

import (
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
)

//go:generate mockgen -source werkerteller.go -destination werkerteller.mock.go -package build_signaler

type CommitPushWerkerTeller interface {
	TellWerker(push *pb.Push, conf *Signaler, handler models.VCSHandler, token string, force bool, sigBy pb.SignaledBy) error
}

type PRWerkerTeller interface {
	TellWerker(push *pb.PullRequest, conf *Signaler, handler models.VCSHandler, token string, force bool, sigBy pb.SignaledBy) error
}

func BuildChangesetData(handler models.VCSHandler, acctRepo, latestCommit, branch string, commits []*pb.Commit) (*pb.ChangesetData, error){
	if len(commits) == 0 {
		return nil, nil
	}
	commitLen := len(commits)
	commitMsgs := make([]string, commitLen)
	earliestCommit := commits[commitLen-1]
	changedFiles, err := handler.GetChangedFiles(acctRepo, latestCommit, earliestCommit.Hash)
	if err != nil {
		return nil, err
	}
	for i, commit := range commits {
		commitMsgs[i] = commit.Message
	}
	return &pb.ChangesetData{FilesChanged: changedFiles, CommitTexts: commitMsgs, Branch: branch}, nil
}
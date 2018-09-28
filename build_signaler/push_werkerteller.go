package build_signaler

import (
	"github.com/pkg/errors"

	ocelog "github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"

	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
)


type PushWerkerTeller struct {}

// TellWerker will get config off of the HeadCommit Hash, Build the werker task object based off of the push head commit data,
// retrieve and generate changeset data from bitbucket to get the files changes / commit messages in the push, then it will add all of that
// to nsq so a werker node can swoop by and pick it up
func (pwt *PushWerkerTeller) TellWerker(push *pb.Push, conf *Signaler, handler models.VCSHandler, token string, force bool, sigBy pb.SignaledBy) (err error) {
	ocelog.Log().WithField("current HEAD hash", push.HeadCommit.Hash).WithField("acctRepo", push.Repo.AcctRepo).WithField("branch", push.Branch).Infof("new build from push of type %s coming in", sigBy.String())
	if token == "" {
		return errors.New("token cannot be empty")
	}
	var buildConf *pb.BuildConfig
	buildConf, err = GetConfig(push.Repo.AcctRepo, push.HeadCommit.Hash, conf.Deserializer, handler)
	if err != nil {
		if err == ocenet.FileNotFound {
			ocelog.IncludeErrField(err).Error("no ocelot.yml")
			return errors.New("no ocelot yaml found for repo " + push.Repo.AcctRepo)
		}
		ocelog.IncludeErrField(err).Error("couldn't get ocelot.yml")
		return errors.Wrap(err, "unable to get build configuration")
	}
	task := BuildInitialWerkerTask(buildConf, push.HeadCommit.Hash, token, push.Branch, push.Repo.AcctRepo, sigBy, nil)
	if push.PreviousHeadCommit == nil {
		task.ChangesetData, err = GenerateNoPreviousHeadChangeset(handler, push.Repo.AcctRepo, push.Branch, push.HeadCommit.Hash)
	} else {
		task.ChangesetData, err = GenerateChangesetFromVCS(handler, push.Repo.AcctRepo, push.Branch, push.HeadCommit.Hash, push.PreviousHeadCommit.Hash, push.Commits)
	}
	if err != nil {
		return errors.Wrap(err, "did not queue because unable to contact vcs repo to get changelist data")
	}
	if err = conf.CheckViableThenQueueAndStore(task, force, push.Commits); err != nil {
		if _, ok := err.(*build.NotViable); ok {
			return errors.New("did not queue because it shouldn't be queued. explanation: " + err.Error())
		}
		ocelog.IncludeErrField(err).Warn("something went awry trying to queue and store")
		return errors.Wrap(err, "unable to queue or store")
	}
	return nil
}

// GenerateChangeSetFromVCS will retrieve and fill out fields of a Changeset by calling the handler for the diff of the latest Commit Hash and earliest Commit hash
//  and by pulling out all commit messages from the commit list. these will be used as data to check criteria of triggers
func GenerateChangesetFromVCS(handler models.VCSHandler, acctRepo, branch, latestCommitHash, earliestCommitHash string, commits []*pb.Commit) (*pb.ChangesetData, error) {
	var commitMsgs []string
	changedFiles, err := handler.GetChangedFiles(acctRepo, latestCommitHash, earliestCommitHash)
	if err != nil {
		return nil, err
	}
	for _, commit := range commits {
		commitMsgs = append(commitMsgs, commit.Message)
	}
	return &pb.ChangesetData{FilesChanged: changedFiles, CommitTexts: commitMsgs, Branch: branch}, nil

}


func GenerateNoPreviousHeadChangeset(handler models.VCSHandler, acctRepo, branch, latestCommitHash string) (*pb.ChangesetData, error) {
	changedFiles, err := handler.GetChangedFiles(acctRepo, latestCommitHash, "")
	if err != nil {
		return nil, err
	}
	return &pb.ChangesetData{Branch: branch, FilesChanged: changedFiles}, nil
}
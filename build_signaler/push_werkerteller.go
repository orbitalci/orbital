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
	// todo: change build changeset data to also take in first / last hash for diffing instead of taking first/last in commit array.
	task.ChangesetData, err = BuildChangesetData(handler, push.Repo.AcctRepo, push.HeadCommit.Hash, push.Branch, push.Commits)
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
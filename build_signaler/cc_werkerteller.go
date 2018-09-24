package build_signaler

import (
	"github.com/pkg/errors"

	ocelog "github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"

	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
)

type CCWerkerTeller struct{}



func (w *CCWerkerTeller) TellWerker(hash string, signaler *Signaler, branch string, handler models.VCSHandler, token, acctRepo string, commits []*pb.Commit, force bool, sigType pb.SignaledBy) (err error) {
	ocelog.Log().WithField("hash", hash).WithField("acctRepo", acctRepo).WithField("branch", branch).Info("found new commit")
	if token == "" {
		return errors.New("token cannot be empty")
	}
	var buildConf *pb.BuildConfig
	buildConf, err = GetConfig(acctRepo, hash, signaler.Deserializer, handler)
	if err != nil {
		if err == ocenet.FileNotFound {
			ocelog.IncludeErrField(err).Error("no ocelot.yml")
			return errors.New("no ocelot yaml found for repo " + acctRepo)
		}
		ocelog.IncludeErrField(err).Error("couldn't get ocelot.yml")
		return errors.Wrap(err, "unable to get build configuration")
	}
	task := BuildInitialWerkerTask(buildConf, hash, token, branch, acctRepo, sigType, nil)
	task.ChangesetData, err = BuildChangesetData(handler, acctRepo, hash, branch, commits)
	if err != nil {
		return errors.Wrap(err, "did not queue because unable to contact vcs repo to get changelist data")
	}
	if err = signaler.CheckViableThenQueueAndStore(task, force, commits); err != nil {
		if _, ok := err.(*build.NotViable); ok {
			return errors.New("did not queue because it shouldn't be queued. explanation: " + err.Error())
		}
		ocelog.IncludeErrField(err).Warn("something went awry trying to queue and store")
		return errors.Wrap(err, "unable to queue or store")
	}
	return nil
}

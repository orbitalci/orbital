package webhook

import (
	"github.com/pkg/errors"

	"github.com/shankj3/ocelot/build"
	signal "github.com/shankj3/ocelot/build_signaler"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"

	ocelog "github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"
)

//GetPrWerkerTeller will return a new PRWerkerTeller object with the private fields prId and destBranch instantiated
func GetPrWerkerTeller(prdata *pb.PrWerkerData, destBranch string) *PRWerkerTeller {
	return &PRWerkerTeller{
		prData:     prdata,
		destBranch: destBranch,
	}
}

//PRWerkerTeller is a queuing struct for pull requests. Since pull requests build triggers will need to validate if
// the build is viable by the *destination* branch, but everything else be queued off the *source* details,
// there needs to be some special finegaling.
type PRWerkerTeller struct {
	destBranch string
	prData     *pb.PrWerkerData
}

//TellWerker will get the ocelot.yml configuration, build the werker task, validate viablilty off of the (*PRWerkerTeller).destBranch field instead off the normal branch passed , then will queue the build using the normal passed branch.
func (pr *PRWerkerTeller) TellWerker(hash string, signaler *signal.Signaler, branch string, handler models.VCSHandler, token, acctRepo string, commits []*pb.Commit, force bool, sigBy pb.SignaledBy) error {
	buildConf, err := signal.GetConfig(acctRepo, hash, signaler.Deserializer, handler)
	if err != nil {
		if err == ocenet.FileNotFound {
			ocelog.IncludeErrField(err).Error("no ocelot.yml")
			return errors.New("no ocelot yaml found for repo " + acctRepo)
		} else {
			return errors.Wrap(err, "unable to get build configuration")
		}
	}
	task := signal.BuildInitialWerkerTask(buildConf, hash, token, branch, acctRepo, pb.SignaledBy_PULL_REQUEST, pr.prData)
	task.ChangesetData, err = signal.BuildChangesetData(handler, acctRepo, hash, branch, commits)
	if err != nil {
		return errors.Wrap(err, "did not queue because unable to contact vcs repo to get changelist data")
	}
	err = signaler.OcyValidator.ValidateViability(pr.destBranch, buildConf.Branches, commits, false)
	if err != nil {
		ocelog.IncludeErrField(err).Warn("fyi, this pull request is not valid for a build!! it will not be queued!!")
		return errors.Wrap(err, "did not queue because it shouldn't be queued, as there is a validation error")
	}
	if err = signaler.QueueAndStore(task); err != nil {
		if _, ok := err.(*build.NotViable); ok {
			ocelog.IncludeErrField(err).Warn("fyi, this pull request is not valid for a build!! it will not be queued!!")
			return errors.Wrap(err, "did not queue because it shouldn't be queued, as there is a validation error")
		}
		ocelog.IncludeErrField(err).Warn("something went awry trying to queue and store")
		return errors.Wrap(err, "unable to queue or store")
	}
	return nil
}

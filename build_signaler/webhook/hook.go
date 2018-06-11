package webhook

import (
	"errors"

	"github.com/shankj3/ocelot/build"
	signal "github.com/shankj3/ocelot/build_signaler"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"

	ocelog "github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"
)

//GetPrWerkerTeller will return a new PRWerkerTeller object with the private fields prId and destBranch instantiated
func GetPrWerkerTeller(prId string, destBranch string) *PRWerkerTeller {
	return &PRWerkerTeller{
		prId: prId,
		destBranch: destBranch,
	}
}

//PRWerkerTeller is a queuing struct for pull requests. Since pull requests build triggers will need to validate if
// the build is viable by the *destination* branch, but everything else be queued off the *source* details,
// there needs to be some special finegaling.
type PRWerkerTeller struct {
	prId string
	destBranch string
}

//TellWerker will get the ocelot.yml configuration, build the werker task, validate viablilty off of the (*PRWerkerTeller).destBranch field instead off the normal branch passed , then will queue the build using the normal passed branch.
func (pr *PRWerkerTeller) TellWerker(hash string, signaler *signal.Signaler, branch string, handler models.VCSHandler, token, acctRepo string, commits []*pb.Commit, force bool, sigBy pb.SignaledBy) error {
	buildConf, err := signal.GetConfig(acctRepo, hash, signaler.Deserializer, handler)
	if err != nil {
		if err == ocenet.FileNotFound {
			err := errors.New("no ocelot yml found for repo " + acctRepo)
			ocelog.IncludeErrField(err).Error("no ocelotyml")
			return err
		} else {
			err := errors.New("unable to get build configuration; err: " + err.Error())
			ocelog.IncludeErrField(err).Error("couldn't get ocelotyml")
			return err
		}
	}
	task := signal.BuildInitialWerkerTask(buildConf, hash, token, branch, acctRepo, pb.SignaledBy_PULL_REQUEST, pr.prId)
	err = signaler.OcyValidator.ValidateViability(pr.destBranch, buildConf.Branches, commits, false)
	if err != nil {
		ocelog.IncludeErrField(err).Warn("fyi, this pull request is not valid for a build!! it will not be queued!!")
		return err
	}
	if err = signaler.QueueAndStore(task); err != nil {
		if _, ok := err.(*build.NotViable); ok {
			ocelog.IncludeErrField(err).Warn("fyi, this pull request is not valid for a build!! it will not be queued!!")
			return err
		}
		ocelog.IncludeErrField(err).Warn("something went awry trying to queue and store")
	}
	return nil
}
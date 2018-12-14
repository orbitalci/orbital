package webhook

import (
	"github.com/pkg/errors"

	"github.com/level11consulting/orbitalci/build/buildeventhandler/push/buildjob"
	"github.com/level11consulting/orbitalci/build/helpers/buildscript/download"
	"github.com/level11consulting/orbitalci/client/buildconfigvalidator"
	"github.com/level11consulting/orbitalci/models"
	"github.com/level11consulting/orbitalci/models/pb"

	ocelog "github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"
)

type PullReqWerkerTeller struct{}

// TellWerker will use the pullRequest COMMITS url to retrieve all commits associated with
func (pr *PullReqWerkerTeller) TellWerker(pullreq *pb.PullRequest, prData *pb.PrWerkerData, signaler *buildjob.Signaler, handler models.VCSHandler, token string, force bool, sigBy pb.SignaledBy) error {
	buildConf, err := download.GetConfig(pullreq.Source.Repo.AcctRepo, pullreq.Source.Hash, signaler.Deserializer, handler)
	if err != nil {
		if err == ocenet.FileNotFound {
			ocelog.IncludeErrField(err).Error("no ocelot.yml")
			return errors.New("no ocelot yaml found for repo " + pullreq.Source.Repo.AcctRepo)
		}
		ocelog.IncludeErrField(err).Error("couldn't get ocelot.yml")
		return errors.Wrap(err, "unable to get build configuration")

	}
	task := buildjob.BuildInitialWerkerTask(buildConf, pullreq.Source.Hash, token, pullreq.Source.Branch, pullreq.Source.Repo.AcctRepo, pb.SignaledBy_PULL_REQUEST, prData, handler.GetVcsType())

	// todo: right now, we do not want to build out the changed files and commit messages, because those should just be done on push events
	// maybe we should re-evaluate down the road? idk, just seems like duping that work if its done on push & pr...
	task.ChangesetData = &pb.ChangesetData{Branch: pullreq.Source.Branch}

	// we don't want to use all the commits to validate whether or not to build this. one commit in the whole branch / fork before submitting a PR might have had CI SKIP, but
	// that doesn't mean the pr shouldn't be built, just that push that had the comment w/ CI SKIP
	//commits, err := handler.GetPRCommits(pullreq.GetUrls().GetCommits())
	//if err != nil {
	//	ocelog.IncludeErrField(err).Error("unable to get pr commits")
	//	return errors.Wrap(err, "unable to get pr commits")
	//}
	err = signaler.OcyValidator.ValidateViability(pullreq.Destination.Branch, buildConf.Branches, []*pb.Commit{}, false)
	if err != nil {
		ocelog.IncludeErrField(err).Warn("fyi, this pull request is not valid for a build!! it will not be queued!!")
		return errors.Wrap(err, "did not queue because it shouldn't be queued, as there is a validation error")
	}
	if err = signaler.QueueAndStore(task); err != nil {
		if _, ok := err.(*buildconfigvalidator.NotViable); ok {
			ocelog.IncludeErrField(err).Warn("fyi, this pull request is not valid for a build!! it will not be queued!!")
			return buildconfigvalidator.NoViability(errors.Wrap(err, "did not queue because it shouldn't be queued, as there is a validation error").Error())
		}
		ocelog.IncludeErrField(err).Warn("something went awry trying to queue and store")
		return errors.Wrap(err, "unable to queue or store")
	}
	return nil
}

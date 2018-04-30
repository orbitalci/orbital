package poll

import (
	"errors"

	ocelog "github.com/shankj3/go-til/log"
	signal "github.com/shankj3/ocelot/build_signaler"
	"github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/common/remote"
	"github.com/shankj3/ocelot/models"
)

type ChangeChecker struct {
	*signal.Signaler
	handler models.VCSHandler
	token   string
}

func (w *ChangeChecker) SetAuth() error {
	cfg, err := credentials.GetVcsCreds(w.Store, w.AcctRepo, w.RC)
	if err != nil {
		return errors.New("couldn't get vcs creds, error: " + err.Error())
	}
	handler, token, err := remote.GetHandler(cfg)
	if err != nil {
		return errors.New("could not get remote client, error: " + err.Error())
	}
	w.handler = handler
	w.token = token
	return nil
}

func (w *ChangeChecker) InspectCommits(branch string, lastHash string) (newLastHash string, err error) {
	commits, err := w.handler.GetAllCommits(w.AcctRepo, branch)
	if err != nil {
		return "", errors.New("could not get all commits, error: " + err.Error())
	}
	if len(commits.Values) == 0 {
		return "", errors.New("no commits found; likely a branch misconfiguration")
	}
	lastCommit := commits.Values[0]
	wt := &signal.BBWerkerTeller{}
	//lastCommitDt := time.Unix(lastCommit.Date.Seconds, int64(lastCommit.Date.Nanos))
	// check for empty last hash now that you have the last commit info and can trigger a build
	if lastHash == "" {
		newLastHash = lastCommit.Hash
		if err = wt.TellWerker(lastCommit.Hash, w.Signaler, branch, w.handler, w.token); err != nil {
			ocelog.IncludeErrField(err).Error("could not queue!")
		}
		return
	}
	//ocelog.Log().WithField("lastCommitDt", lastCommitDt.String()).Info()
	if lastHash != lastCommit.Hash {
		ocelog.Log().Infof("found a new hash %s, telling werker", lastCommit.Hash)
		newLastHash = lastCommit.Hash
		if err = wt.TellWerker(lastCommit.Hash, w.Signaler, branch, w.handler, w.token); err != nil {
			return
		}
	} else {
		// nothing happened, have the hash map reflect that. set the "new" value to the old one, and mosey along.
		newLastHash = lastHash
		ocelog.Log().WithField("acctRepo", w.AcctRepo).WithField("branch", branch).Info("no new commits found")
	}
	return
}

func searchBranchCommits(handler models.VCSHandler, branch string, conf *ChangeChecker, lastHash string, token string, wt signal.WerkerTeller) (newLastHash string, err error) {
	commits, err := handler.GetAllCommits(conf.AcctRepo, branch)
	if err != nil {
		ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Error("couldn't get commits ")
		return
	}
	if len(commits.Values) == 0 {
		ocelog.Log().Fatal("no commits found. likely a branch misconfiguration. exiting.")
	}
	lastCommit := commits.Values[0]
	//lastCommitDt := time.Unix(lastCommit.Date.Seconds, int64(lastCommit.Date.Nanos))
	// check for empty last hash now that you have the last commit info and can trigger a build
	if lastHash == "" {
		newLastHash = lastCommit.Hash
		ocelog.Log().Info("there was no lastHash entry in the map, so running a build off of the latest commit")
		if err = wt.TellWerker(lastCommit.Hash, conf.Signaler, branch, handler, token); err != nil {
			ocelog.IncludeErrField(err).Error("could not queue!")
		}
		return
	}
	//ocelog.Log().WithField("lastCommitDt", lastCommitDt.String()).Info()
	if lastHash != lastCommit.Hash {
		ocelog.Log().Infof("found a new hash %s, telling werker", lastCommit.Hash)
		newLastHash = lastCommit.Hash
		if err = wt.TellWerker(lastCommit.Hash, conf.Signaler, branch, handler, token); err != nil {
			return
		}
	} else {
		// nothing happened, have the hash map reflect that. set the "new" value to the old one, and mosey along.
		newLastHash = lastHash
		ocelog.Log().WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Info("no new commits found")
	}
	return
}

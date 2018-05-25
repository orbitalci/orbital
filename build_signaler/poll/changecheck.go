package poll

import (
	"errors"
	"time"

	ocelog "github.com/shankj3/go-til/log"
	signal "github.com/shankj3/ocelot/build_signaler"
	"github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/common/remote"
	"github.com/shankj3/ocelot/models"
)

func NewChangeChecker(signaler *signal.Signaler) *ChangeChecker {
	return &ChangeChecker{
		Signaler:signaler,
		teller: &signal.BBWerkerTeller{},
	}
}

type ChangeChecker struct {
	*signal.Signaler
	handler models.VCSHandler
	token   string
	teller  signal.WerkerTeller
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

func (w *ChangeChecker) HandleAllBranches(branchLastHashes map[string]string) error {
	branchHistories, err := w.handler.GetAllBranchesLastCommitData(w.AcctRepo)
	if err != nil {
		return errors.New("could not get branch history, error is: " + err.Error())
	}
	if len(branchHistories) == 0 {
		return errors.New("no branches found, likely an acct/repo misconfiguration")
	}
	for _, branchHist := range branchHistories {
		lastHash, ok := branchLastHashes[branchHist.Branch]
		if ok {
			if lastHash != branchHist.Hash {
				err = w.teller.TellWerker(branchHist.Hash, w.Signaler, branchHist.Branch, w.handler, w.token)
				branchLastHashes[branchHist.Branch] = branchHist.Hash
				if err != nil {
					return err
				}
			}
		} else {
			// add to map so we can track this branch
			branchLastHashes[branchHist.Branch] = branchHist.Hash
			// this has never been built/tracked before... so if anything has been committed in the last week, build it and add it to the map
			lastCommitTime := time.Unix(branchHist.LastCommitTime.Seconds, int64(branchHist.LastCommitTime.Nanos))
			lastWeek := time.Now().Add(-time.Hour*24*7)
			if lastWeek.After(lastCommitTime) {
				if err = w.teller.TellWerker(branchHist.Hash, w.Signaler, branchHist.Branch, w.handler, w.token); err != nil {
					return err
				}
			}
		}
	}
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
	//lastCommitDt := time.Unix(lastCommit.Date.Seconds, int64(lastCommit.Date.Nanos))
	// check for empty last hash now that you have the last commit info and can trigger a build
	if lastHash == "" {
		newLastHash = lastCommit.Hash
		if err = w.teller.TellWerker(lastCommit.Hash, w.Signaler, branch, w.handler, w.token); err != nil {
			ocelog.IncludeErrField(err).Error("could not queue!")
		}
		return
	}
	//ocelog.Log().WithField("lastCommitDt", lastCommitDt.String()).Info()
	if lastHash != lastCommit.Hash {
		ocelog.Log().Infof("found a new hash %s, telling werker", lastCommit.Hash)
		newLastHash = lastCommit.Hash
		if err = w.teller.TellWerker(lastCommit.Hash, w.Signaler, branch, w.handler, w.token); err != nil {
			return
		}
	} else {
		// nothing happened, have the hash map reflect that. set the "new" value to the old one, and mosey along.
		newLastHash = lastHash
		ocelog.Log().WithField("acctRepo", w.AcctRepo).WithField("branch", branch).Info("no new commits found")
	}
	return
}

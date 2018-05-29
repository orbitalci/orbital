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

// SetAuth retrieves VCS credentials based on the account, then creates a VCS handler with it.
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

// HandleAllBranches will get all branches associated with the repository along with their last commit information.
//  If the branch is already in the map and the commit in the branch map is different than the one retrieved from bitbucket,
//    then a build will be triggered ad teh branchLastHashes map will be updated with the newest commit
//
//  If the branch is not in the map, and if there is a new commit on the branch in the last week, a build will be triggered.
//    If the branch is not in the map but there are no recent commits, then the map will be updated to include this branch,
//    but a build will not be triggered.
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
			ocelog.Log().Info("this branch is already being tracked, checking if the built hash is the same as the one retrieved from bitbucket")
			if lastHash != branchHist.Hash {
				ocelog.Log().Info("hashes are not the same, telling werker...")
				err = w.teller.TellWerker(branchHist.Hash, w.Signaler, branchHist.Branch, w.handler, w.token)
				branchLastHashes[branchHist.Branch] = branchHist.Hash
				if err != nil {
					return err
				}
			}
		} else {
			ocelog.Log().Info("branch was not previously tracked by ocelot, checking if its worthy of build")
			// add to map so we can track this branch
			branchLastHashes[branchHist.Branch] = branchHist.Hash
			// this has never been built/tracked before... so if anything has been committed in the last week, build it and add it to the map
			lastCommitTime := time.Unix(branchHist.LastCommitTime.Seconds, int64(branchHist.LastCommitTime.Nanos))
			lastWeek := time.Now().Add(-time.Hour*24*7)
			if lastWeek.After(lastCommitTime) {
				ocelog.Log().Info("it is! it has been active at least in the past week, it will be built then added to ocelot tracking")
				if err = w.teller.TellWerker(branchHist.Hash, w.Signaler, branchHist.Branch, w.handler, w.token); err != nil {
					return err
				}
			} else {
				ocelog.Log().Info("it is not! adding branch to tracking list, but not telling werker")
			}
		}
	}
	ocelog.Log().Info("finished checking all branches")
	return nil
}

// InspectCommits will retrieve the last commit data for a branch of a repository. If the value of lastHash is empty,
//   then a build will be triggered for this branch. If lastHash is not empty, and the retrieved last commit hash is not
//   equal to the lastHash passed, then a build will be triggered. If the hashes are equal, then no build will be triggered.
func (w *ChangeChecker) InspectCommits(branch string, lastHash string) (newLastHash string, err error) {
	lastCommit, err := w.handler.GetBranchLastCommitData(w.AcctRepo, branch)
	if err != nil {
		return "", errors.New("could not get all commits, error: " + err.Error())
	}
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

package poll

import (
	"errors"
	"time"

	"github.com/level11consulting/ocelot/build/eventhandler/push/buildjob"
	"github.com/level11consulting/ocelot/build/vcshandler"
	signal "github.com/level11consulting/ocelot/build_signaler"
	"github.com/level11consulting/ocelot/client/newbuildjob"
	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/server/config"
	ocelog "github.com/shankj3/go-til/log"
)

func NewChangeChecker(signaler *buildjob.Signaler, acctRepo string, vcsType pb.SubCredType) *ChangeChecker {
	return &ChangeChecker{
		Signaler: signaler,
		AcctRepo: acctRepo,
		pTeller:  &newbuildjob.PushWerkerTeller{},
		vcsType:  vcsType,
	}
}

type ChangeChecker struct {
	*buildjob.Signaler
	handler  models.VCSHandler
	token    string
	AcctRepo string
	vcsType  pb.SubCredType
	pTeller  signal.CommitPushWerkerTeller
}

// SetAuth retrieves VCS credentials based on the account, then creates a VCS handler with it.
func (w *ChangeChecker) SetAuth() error {
	cfg, err := config.GetVcsCreds(w.Store, w.AcctRepo, w.RC, w.vcsType)
	if err != nil {
		return errors.New("couldn't get vcs creds, error: " + err.Error())
	}
	handler, token, err := vcshandler.GetHandler(cfg)
	if err != nil {
		return errors.New("could not get remote client, error: " + err.Error())
	}
	w.handler = handler
	w.token = token
	return nil
}

func (w *ChangeChecker) generatePush(acctRepo, previousHash, latestHash, branch string) (*pb.Push, error) {
	var previousHeadCommit *pb.Commit
	var newHeadCommit *pb.Commit
	var pushcommits []*pb.Commit
	if previousHash != "" {
		commits, err := w.generateCommitList(acctRepo, branch, previousHash)
		if err != nil {
			return nil, err
		}
		commitNum := len(commits)
		if commitNum == 0 {
			return nil, errors.New("no commits in push, nothing to do")
		}
		previousHeadCommit = commits[commitNum-1]
		pushcommits = commits[0 : commitNum-1]
		newHeadCommit = commits[0]
	} else {
		previousHeadCommit = nil
		pushcommits = []*pb.Commit{}
		newHeadCommit = &pb.Commit{Hash: latestHash}
	}
	return &pb.Push{
		Commits:            pushcommits,
		Branch:             branch,
		Repo:               &pb.Repo{AcctRepo: acctRepo},
		HeadCommit:         newHeadCommit,
		PreviousHeadCommit: previousHeadCommit,
	}, nil
}

// generateCommitList just calls the handler function to get commit log. should be mirrored in hook.go
func (w *ChangeChecker) generateCommitList(acctRepo string, branch string, lastHash string) (commits []*pb.Commit, err error) {
	commits, err = w.handler.GetCommitLog(acctRepo, branch, lastHash)
	if err != nil {
		commits = nil
		ocelog.IncludeErrField(err).Error("unable to get commit list from VCS handler!! oh nuuu")
	}
	return
}

// HandleAllBranches will get all branches associated with the repository along with their last commit information.
//  If the branch is already in the map and the commit in the branch map is different than the one retrieved from bitbucket,
//    then a build will be triggered ad teh branchLastHashes map will be updated with the newest commit. A full changeset to
//    trigger certain stages on will also be available the will be the files changed between the last commit and the most recent one, and
//    all the commit messages
//
//  If the branch is not in the map, and if there is a new commit on the branch in the last week, a build will be triggered.
//    This type of trigger will not result in a changeset of files changed / commit messages to trigger stages on.
//  If the branch is not in the map but there are no recent commits, then the map will be updated to include this branch,
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
			ocelog.Log().WithField("branch", branchHist.Branch).Info("this branch is already being tracked, checking if the built hash is the same as the one retrieved from VCS ")
			if lastHash != branchHist.Hash {
				ocelog.Log().
					WithField("branch", branchHist.Branch).
					WithField("db last hash", lastHash).
					WithField("most recent hash", branchHist.Hash).
					Info("hashes are not the same, telling werker...")
				writtenPush, err := w.generatePush(w.AcctRepo, lastHash, "", branchHist.Branch)
				if err != nil {
					return err
				}
				ocelog.Log().WithField("generated push lash hash", writtenPush.PreviousHeadCommit.Hash).Info()
				err = w.pTeller.TellWerker(writtenPush, w.Signaler, w.handler, w.token, false, pb.SignaledBy_POLL)
				branchLastHashes[branchHist.Branch] = branchHist.Hash
				if err != nil {
					return err
				}
			}
		} else {
			ocelog.Log().WithField("branch", branchHist.Branch).Info("branch was not previously tracked by ocelot, checking if its worthy of build")
			// add to map so we can track this branch
			branchLastHashes[branchHist.Branch] = branchHist.Hash
			// this has never been built/tracked before... so if anything has been committed in the last week, build it and add it to the map
			lastCommitTime := time.Unix(branchHist.LastCommitTime.Seconds, int64(branchHist.LastCommitTime.Nanos))
			lastWeek := time.Now().AddDate(0, 0, -7)
			ocelog.Log().WithField("last commit time", lastCommitTime.Format("Jan 2 15:04:05 2006")).WithField("last week", lastWeek.Format("Jan 2 15:04:05 2006")).Info("times!")
			if lastCommitTime.After(lastWeek) {
				ocelog.Log().WithField("branch", branchHist.Branch).WithField("hash", branchHist.Hash).Info("it is! it has been active at least in the past week, it will be built then added to ocelot tracking")
				commits, err := w.generatePush(w.AcctRepo, "", branchHist.Hash, branchHist.Branch)
				if err != nil {
					return err
				}
				// since this has never been built before, we aren't going to parse the commit list to check for CI SKIP, we wouldn't have anything to check against
				if err = w.pTeller.TellWerker(commits, w.Signaler, w.handler, w.token, false, pb.SignaledBy_POLL); err != nil {
					return err
				}
			} else {
				ocelog.Log().WithField("branch", branchHist.Branch).Info("it is not! adding branch to tracking list, but not telling werker")
			}
		}
	}
	ocelog.Log().WithField("acctRepo", w.AcctRepo).Info("finished checking all branches")
	return nil
}

// InspectCommits will retrieve the last commit data for a branch of a repository. If the value of lastHash is empty,
//   then a build will be triggered for this branch. If lastHash is not empty, and the retrieved last commit hash is not
//   equal to the lastHash passed, then a build will be triggered. If the hashes are equal, then no build will be triggered.
func (w *ChangeChecker) InspectCommits(branch string, lastBuiltHash string) (newLastHash string, err error) {
	lastCommit, err := w.handler.GetBranchLastCommitData(w.AcctRepo, branch)
	if err != nil {
		return "", errors.New("could not get all commits, error: " + err.Error())
	}
	// check for empty last hash now that you have the last commit info and can trigger a build
	if lastBuiltHash == "" {
		newLastHash = lastCommit.Hash
		commits, err1 := w.generatePush(w.AcctRepo, "", lastCommit.Hash, branch)
		if err1 != nil {
			return "", err1
		}
		// no last hash, therefore not going to check for ci skip
		if err = w.pTeller.TellWerker(commits, w.Signaler, w.handler, w.token, false, pb.SignaledBy_POLL); err != nil {
			ocelog.IncludeErrField(err).Error("could not queue!")
		}
		return
	}
	//ocelog.Log().WithField("lastCommitDt", lastCommitDt.String()).Info()
	if lastBuiltHash != lastCommit.Hash {
		ocelog.Log().Infof("found a new hash %s, telling werker", lastCommit.Hash)
		// this has been tracked before, and we have a last hash to get a commit list so we can check for ci skip
		var generatedPush *pb.Push
		generatedPush, err = w.generatePush(w.AcctRepo, lastBuiltHash, "", branch)
		if err != nil {
			return "", err
		}
		newLastHash = lastCommit.Hash
		if err = w.pTeller.TellWerker(generatedPush, w.Signaler, w.handler, w.token, false, pb.SignaledBy_POLL); err != nil {
			return
		}
	} else {
		// nothing happened, have the hash map reflect that. set the "new" value to the old one, and mosey along.
		newLastHash = lastBuiltHash
		ocelog.Log().WithField("AcctRepo", w.AcctRepo).WithField("branch", branch).Info("no new commits found")
	}
	return
}

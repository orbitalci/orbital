package poll

import "bitbucket.org/level11consulting/ocelot/util/handler"

// made this interface for easy testing
type werkerTeller interface {
	tellWerker(lastCommit *pb.Commit, conf *changeSetConfig, branch string, store storage.OcelotStorage, handler handler.VCSHandler, token string) (err error)
}

type aWerkerTeller struct {}

func (w *aWerkerTeller) tellWerker(lastCommit *pb.Commit, conf *changeSetConfig, branch string, store storage.OcelotStorage, handler handler.VCSHandler, token string) (err error){
	ocelog.Log().WithField("hash", lastCommit.Hash).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Info("found new commit")
	ocelog.Log().WithField("hash", lastCommit.Hash).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Info("getting bitbucket commit")
	var buildConf *pb.BuildConfig
	buildConf, _, err = build.GetBBConfig(conf.RemoteConf, store, conf.AcctRepo, lastCommit.Hash, conf.Deserializer, handler)
	if err != nil {
		ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Error("couldn't get build configuration")
		return
	}
	if err = build.QueueAndStore(lastCommit.Hash, branch, conf.AcctRepo, token, conf.RemoteConf, buildConf, conf.OcyValidator, conf.Producer, store); err != nil {
		ocelog.IncludeErrField(err).WithField("hash", lastCommit.Hash).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Fatal("couldn't add to build queue or store in db")
		return
	}
	ocelog.Log().WithField("hash", lastCommit.Hash).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Info("successfully added build to build queue")
	return
}

func searchBranchCommits(handler handler.VCSHandler, branch string, conf *changeSetConfig, lastHash string, store storage.OcelotStorage, token string, wt werkerTeller) (newLastHash string, err error) {
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
		if err = wt.tellWerker(lastCommit, conf, branch, store, handler, token); err != nil {
			ocelog.IncludeErrField(err).Error("could not queue!")
		}
		return
	}
	//ocelog.Log().WithField("lastCommitDt", lastCommitDt.String()).Info()
	if lastHash != lastCommit.Hash {
		ocelog.Log().Infof("found a new hash %s, telling werker", lastCommit.Hash)
		newLastHash = lastCommit.Hash
		if err = wt.tellWerker(lastCommit, conf, branch, store, handler, token); err != nil {
			return
		}
	} else {
		// nothing happened, have the hash map reflect that. set the "new" value to the old one, and mosey along.
		newLastHash = lastHash
		ocelog.Log().WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Info("no new commits found")
	}
	return
}


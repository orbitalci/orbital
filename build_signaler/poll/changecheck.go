package poll

import (
	"errors"
	
	"bitbucket.org/level11consulting/go-til/deserialize"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	"bitbucket.org/level11consulting/ocelot/build"
	signal "bitbucket.org/level11consulting/ocelot/build_signaler"
	"bitbucket.org/level11consulting/ocelot/common/credentials"
	"bitbucket.org/level11consulting/ocelot/common/remote"
	"bitbucket.org/level11consulting/ocelot/common/remote/bitbucket"
	pbb "bitbucket.org/level11consulting/ocelot/models/bitbucket/pb"
	"bitbucket.org/level11consulting/ocelot/models/pb"
	"bitbucket.org/level11consulting/ocelot/storage"
)

type ChangeChecker struct {
	RC credentials.CVRemoteConfig
	*deserialize.Deserializer
	Producer     *nsqpb.PbProduce
	OcyValidator *build.OcelotValidator
	Store        storage.OcelotStorage
	AcctRepo  	  string
	bbHandler    remote.VCSHandler
	token        string
}

// made this interface for easy testing
type werkerTeller interface {
	tellWerker(lastCommit *pbb.Commit, conf *ChangeChecker, branch string, store storage.OcelotStorage, handler remote.VCSHandler, token string) (err error)
}

type aWerkerTeller struct {}

func (w *aWerkerTeller) tellWerker(lastCommit *pbb.Commit, conf *ChangeChecker, branch string, store storage.OcelotStorage, handler remote.VCSHandler, token string) (err error){
	ocelog.Log().WithField("hash", lastCommit.Hash).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Info("found new commit")
	var buildConf *pb.BuildConfig
	buildConf, _, err = signal.GetBBConfig(conf.RC, store, conf.AcctRepo, lastCommit.Hash, conf.Deserializer, handler)
	if err != nil {
		ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Error("couldn't get build configuration")
		return
	}
	if err = signal.QueueAndStore(lastCommit.Hash, branch, conf.AcctRepo, token, conf.RC, buildConf, conf.OcyValidator, conf.Producer, store); err != nil {
		ocelog.IncludeErrField(err).WithField("hash", lastCommit.Hash).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Fatal("couldn't add to build queue or store in db")
		return
	}
	ocelog.Log().WithField("hash", lastCommit.Hash).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Info("successfully added build to build queue")
	return
}

func (w *ChangeChecker) SetAuth() error {
	cfg, err := credentials.GetVcsCreds(w.Store, w.AcctRepo, w.RC)
	if err != nil {
		return errors.New("couldn't get vcs creds, error: " + err.Error())
	}
	bbHandler, token, err := bitbucket.GetBitbucketClient(cfg)
	if err != nil {
		return errors.New("could not get bitbucket client, error: " + err.Error())
	}
	w.bbHandler = bbHandler
	w.token = token
	return nil
}

func (w *ChangeChecker) InspectCommits(branch string, lastHash string) (newLastHash string, err error) {
	commits, err := w.bbHandler.GetAllCommits(w.AcctRepo, branch)
	if err != nil {
		return "", errors.New("could not get all commits, error: " + err.Error())
	}
	if len(commits.Values) == 0 {
		return "", errors.New("no commits found; likely a branch misconfiguration")
	}
	lastCommit := commits.Values[0]
	wt := &aWerkerTeller{}
	//lastCommitDt := time.Unix(lastCommit.Date.Seconds, int64(lastCommit.Date.Nanos))
	// check for empty last hash now that you have the last commit info and can trigger a build
	if lastHash == "" {
		newLastHash = lastCommit.Hash
		if err = wt.tellWerker(lastCommit, w, branch, w.Store, w.bbHandler, w.token); err != nil {
			ocelog.IncludeErrField(err).Error("could not queue!")
		}
		return
	}
	//ocelog.Log().WithField("lastCommitDt", lastCommitDt.String()).Info()
	if lastHash != lastCommit.Hash {
		ocelog.Log().Infof("found a new hash %s, telling werker", lastCommit.Hash)
		newLastHash = lastCommit.Hash
		if err = wt.tellWerker(lastCommit, w, branch, w.Store, w.bbHandler, w.Token); err != nil {
			return
		}
	} else {
		// nothing happened, have the hash map reflect that. set the "new" value to the old one, and mosey along.
		newLastHash = lastHash
		ocelog.Log().WithField("acctRepo", w.AcctRepo).WithField("branch", branch).Info("no new commits found")
	}
	return
}

func searchBranchCommits(handler remote.VCSHandler, branch string, conf *ChangeChecker, lastHash string, store storage.OcelotStorage, token string, wt werkerTeller) (newLastHash string, err error) {
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


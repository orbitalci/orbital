package main

// needs to:
// receive acct-repo as flag
// call bitbucket for changeset
// check if there have been updates, if there have:
//   - create build message from latest hash
//   - add build message to build topic
// 	 - update last_cron_time in db

import (
	"fmt"
	"os"
	"strings"

	"bitbucket.org/level11consulting/go-til/deserialize"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"bitbucket.org/level11consulting/ocelot/util/build"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/handler"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"github.com/namsral/flag"
)

type changeSetConfig struct {
	RemoteConf   cred.CVRemoteConfig
	*deserialize.Deserializer
	OcyValidator   *build.OcelotValidator
	Producer       *nsqpb.PbProduce
	AcctRepo  	string
	Acct        string
	Repo        string
	Branches     []string
}

func configure() *changeSetConfig {
	var loglevel, consuladdr, acctRepo, branches string
	var consulport int
	flrg := flag.NewFlagSet("poller", flag.ExitOnError)
	flrg.StringVar(&loglevel, "log-level", "info", "log level")
	flrg.StringVar(&acctRepo, "acct-repo", "ERROR", "acct/repo to check changeset for")
	flrg.StringVar(&branches, "branches", "ERROR", "comma separated list of branches to check for changesets")
	flrg.StringVar(&consuladdr, "consul-host", "localhost", "address of consul")
	flrg.IntVar(&consulport, "consul-port", 8500, "port of consul")
	flrg.Parse(os.Args[1:])
	ocelog.InitializeLog(loglevel)
	ocelog.Log().Debug()
	rc, err := cred.GetInstance(consuladdr, consulport, "")
	if err != nil {
		ocelog.IncludeErrField(err).Fatal("unable to get instance of remote config, exiting")
	}
	if acctRepo == "ERROR" || branches == "ERROR" {
		ocelog.Log().Fatal("-acct-repo and -branches is required")
	}
	branchList := strings.Split(branches, ",")
	conf := &changeSetConfig{RemoteConf: rc, AcctRepo: acctRepo, Branches:branchList, Deserializer: deserialize.New(), Producer: nsqpb.GetInitProducer(), OcyValidator: build.GetOcelotValidator()}
	acctrepolist := strings.Split(acctRepo, "/")
	if len(acctrepolist) != 2 {
		ocelog.Log().Fatal("-acct-repo must be in format <acct>/<repo>")
	}
	conf.Acct, conf.Repo = acctrepolist[0], acctrepolist[1]
	return conf
}
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


func main() {
	conf := configure()
	var bbHandler handler.VCSHandler
	var token string
	store, err := conf.RemoteConf.GetOcelotStorage()
	if err != nil {
		ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).Fatal("couldn't get storage")
	}
	defer store.Close()
	cfg, err := build.GetVcsCreds(store, conf.AcctRepo, conf.RemoteConf)
	if err != nil {
		ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).Fatal("why")
	}
	bbHandler, token, err = handler.GetBitbucketClient(cfg)
	fmt.Println(token)
	if err != nil {
		ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).Fatal("why")
	}
	_, lastHashes, err := store.GetLastData(conf.AcctRepo)
	if err != nil {
		ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).Error("couldn't get last cron time, setting last cron to 5 minutes ago")
	}
	// no matter what, we are inside the cron job, so we should be updating the db
	defer func(){
		if err = store.SetLastData(conf.Acct, conf.Repo, lastHashes); err != nil {
			ocelog.IncludeErrField(err).Error("unable to set last cron time")
			return
		}
		ocelog.Log().Info("successfully set last cron time")
		return
	}()

	for _, branch := range conf.Branches {
		lastHash, ok := lastHashes[branch]
		if !ok {
			ocelog.Log().Infof("no last hash found for branch %s in lash Hash map, so this branch will build no matter what", branch)
			lastHash = ""
		}
		newLastHash, err := searchBranchCommits(bbHandler, branch, conf, lastHash, store, token, &aWerkerTeller{})
		ocelog.Log().WithField("old last hash", lastHash).WithField("new last hash", newLastHash).Info("git hash data for poll")
		lastHashes[branch] = newLastHash
		if err != nil {
			ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Error("something went wrong")
		}
	}

}
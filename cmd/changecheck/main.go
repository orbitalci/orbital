package main

// needs to:
// receive acct-repo as flag
// call bitbucket for changeset
// check if there have been updates, if there have:
//   - create build message from latest hash
//   - add build message to build topic
// 	 - update last_cron_time in db

import (
	"bitbucket.org/level11consulting/go-til/deserialize"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"bitbucket.org/level11consulting/ocelot/util/build"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/handler"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"fmt"
	"github.com/namsral/flag"
	"os"
	"strings"
	"time"
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

func searchBranchCommits(handler handler.VCSHandler, branch string, conf *changeSetConfig, lastPoll time.Time, store storage.OcelotStorage, token string) (err error) {
	commits, err := handler.GetAllCommits(conf.AcctRepo, branch)
	if err != nil {
		ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Error("couldn't get commits ")
		return
	}
	lastCommit := commits.Values[0]
	lastCommitDt := time.Unix(lastCommit.Date.Seconds, int64(lastCommit.Date.Nanos))
	if lastCommitDt.After(lastPoll) {
		ocelog.Log().WithField("hash", lastCommit.Hash).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Info("found new commit")
		ocelog.Log().WithField("hash", lastCommit.Hash).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Info("getting bitbucket commit")
		var buildConf *pb.BuildConfig
		buildConf, _, err = build.GetBBConfig(conf.RemoteConf, conf.AcctRepo, lastCommit.Hash, conf.Deserializer, handler)
		if err != nil {
			ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Error("couldn't get build configuration")
			return
		}
		if err = build.QueueAndStore(lastCommit.Hash, branch, conf.AcctRepo, token, conf.RemoteConf, buildConf, conf.OcyValidator, conf.Producer, store); err != nil {
			ocelog.IncludeErrField(err).WithField("hash", lastCommit.Hash).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Fatal("couldn't add to build queue or store in db")
			return
		}
		ocelog.Log().WithField("hash", lastCommit.Hash).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Info("successfully added build to build queue")
		if err = store.SetLastCronTime(conf.Acct, conf.Repo); err != nil {
			ocelog.IncludeErrField(err).WithField("hash", lastCommit.Hash).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Error("unable to set last cron time")
			return
		}
		ocelog.Log().WithField("hash", lastCommit.Hash).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Info("successfully set last cron time")
	} else {
		ocelog.Log().WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Info("no new commits found")
	}
	return
}


func main() {
	conf := configure()
	var bbHandler handler.VCSHandler
	var token string
	cfg, err := build.GetVcsCreds(conf.AcctRepo, conf.RemoteConf)
	if err != nil {
		ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).Fatal("why")
	}
	bbHandler, token, err = handler.GetBitbucketClient(cfg)
	fmt.Println(token)
	if err != nil {
		ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).Fatal("why")
	}
	store, err := conf.RemoteConf.GetOcelotStorage()
	if err != nil {
		ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).Fatal("couldn't get storage")
	}
	lastCron, err := store.GetLastCronTime(conf.AcctRepo)
	if err != nil {
		lastCron = time.Now().Add(-5 * time.Minute)
		ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).Error("couldn't get last cron time, setting last cron to 5 minutes ago")
	}
	ocelog.Log().Debug("last cron time is ", lastCron.String())
	ocelog.Log().WithField("lastCronTime", lastCron.String()).Info("checking for new commits")
	for _, branch := range conf.Branches {
		if err := searchBranchCommits(bbHandler, branch, conf, lastCron, store, token); err != nil {
			ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Error("something went wrong")
		}
	}
}
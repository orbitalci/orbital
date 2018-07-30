package main

// needs to:
// receive acct-repo as flag
// call bitbucket for changeset
// check if there have been updates, if there have:
//   - create build message from latest hash
//   - add build message to build topic
// 	 - update last_cron_time in db

import (
	"os"
	"strings"

	"github.com/namsral/flag"
	"github.com/shankj3/go-til/deserialize"
	ocelog "github.com/shankj3/go-til/log"
	"github.com/shankj3/go-til/nsqpb"
	"github.com/shankj3/ocelot/build"
	signal "github.com/shankj3/ocelot/build_signaler"
	"github.com/shankj3/ocelot/build_signaler/poll"
	cred "github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/version"
)

type changeSetConfig struct {
	RemoteConf cred.CVRemoteConfig
	*deserialize.Deserializer
	OcyValidator *build.OcelotValidator
	Producer     *nsqpb.PbProduce
	AcctRepo     string
	Acct         string
	Repo         string
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
	version.MaybePrintVersion(flrg.Args())
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
	conf := &changeSetConfig{RemoteConf: rc, AcctRepo: acctRepo, Branches: branchList, Deserializer: deserialize.New(), Producer: nsqpb.GetInitProducer(), OcyValidator: build.GetOcelotValidator()}
	acctrepolist := strings.Split(acctRepo, "/")
	if len(acctrepolist) != 2 {
		ocelog.Log().Fatal("-acct-repo must be in format <acct>/<repo>")
	}
	conf.Acct, conf.Repo = acctrepolist[0], acctrepolist[1]
	return conf
}

func main() {
	conf := configure()
	store, err := conf.RemoteConf.GetOcelotStorage()
	if err != nil {
		ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).Fatal("couldn't get storage")
	}
	defer store.Close()
	sig := &signal.Signaler{
		RC:           conf.RemoteConf,
		Deserializer: conf.Deserializer,
		Producer:     conf.Producer,
		OcyValidator: conf.OcyValidator,
		Store:        store,
	}
	checker := poll.NewChangeChecker(sig, conf.AcctRepo)

	if err := checker.SetAuth(); err != nil {
		ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).Fatal("could not get auth")
	}

	_, lastHashes, err := store.GetLastData(conf.AcctRepo)
	if err != nil {
		ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).Fatal("couldn't get last hashes")
	}
	// no matter what, we are inside the cron job, so we should be updating the db
	defer func() {
		if err = store.SetLastData(conf.Acct, conf.Repo, lastHashes); err != nil {
			ocelog.IncludeErrField(err).Error("unable to set last cron time")
			return
		}
		ocelog.Log().Info("successfully set last cron time")
		return
	}()
	if len(conf.Branches) == 1 && conf.Branches[0] == "ALL" {
		err = checker.HandleAllBranches(lastHashes)
		if err != nil {
			ocelog.IncludeErrField(err).Error("could not check through branches")
		}
	} else {
		for _, branch := range conf.Branches {
			lastHash, ok := lastHashes[branch]
			if !ok {
				ocelog.Log().Infof("no last hash found for branch %s in lash Hash map, so this branch will build no matter what", branch)
				lastHash = ""
			}
			newLastHash, err := checker.InspectCommits(branch, lastHash)
			if err != nil {
				ocelog.IncludeErrField(err).Error("error searching branch commits, err: " + err.Error())
				return
			}
			ocelog.Log().WithField("old last hash", lastHash).WithField("new last hash", newLastHash).Info("git hash data for poll")
			lastHashes[branch] = newLastHash
			if err != nil {
				ocelog.IncludeErrField(err).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Error("something went wrong")
			}
		}
	}

}

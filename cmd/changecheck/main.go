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

	"net/url"

	"github.com/level11consulting/ocelot/build"
	"github.com/level11consulting/ocelot/build/helpers/stringbuilder"
	signal "github.com/level11consulting/ocelot/build_signaler"
	"github.com/level11consulting/ocelot/build_signaler/poll"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/server/config"
	"github.com/level11consulting/ocelot/version"
	"github.com/namsral/flag"
	"github.com/shankj3/go-til/deserialize"
	ocelog "github.com/shankj3/go-til/log"
	"github.com/shankj3/go-til/nsqpb"
)

type changeSetConfig struct {
	RemoteConf config.CVRemoteConfig
	*deserialize.Deserializer
	OcyValidator *build.OcelotValidator
	Producer     *nsqpb.PbProduce
	AcctRepo     string
	Acct         string
	Repo         string
	Branches     []string
	VcsType      pb.SubCredType
}

// FIXME: consistency: consul's host and port, the var name for configInstance/rc
func configure() *changeSetConfig {
	var loglevel, consuladdr, acctRepo, branches, vcsType string
	var consulport int
	flrg := flag.NewFlagSet("poller", flag.ExitOnError)
	flrg.StringVar(&loglevel, "log-level", "info", "log level")
	flrg.StringVar(&acctRepo, "acct-repo", "ERROR", "acct/repo to check changeset for")
	flrg.StringVar(&branches, "branches", "ERROR", "comma separated list of branches to check for changesets")
	flrg.StringVar(&vcsType, "vcs-type", "ERROR", fmt.Sprintf("%s", strings.Join(pb.CredType_VCS.SubtypesString(), "|")))
	flrg.StringVar(&consuladdr, "consul-host", "localhost", "address of consul")
	flrg.IntVar(&consulport, "consul-port", 8500, "port of consul")
	flrg.Parse(os.Args[1:])
	version.MaybePrintVersion(flrg.Args())
	ocelog.InitializeLog(loglevel)
	ocelog.Log().Debug()

	parsedConsulURL, parsedErr := url.Parse(fmt.Sprintf("consul://%s:%d", consuladdr, consulport))
	if parsedErr != nil {
		ocelog.IncludeErrField(parsedErr).Fatal("failed parsing consul uri, bailing")
	}
	rc, err := config.GetInstance(parsedConsulURL, "")
	if err != nil {
		ocelog.IncludeErrField(err).Fatal("unable to get instance of remote config, exiting")
	}
	if acctRepo == "ERROR" || branches == "ERROR" || vcsType == "ERROR" {
		ocelog.Log().Fatal("-acct-repo, -branches, and -vcs-type are required")
	}
	var ok bool
	var vcsTypeInt int32
	vcsTypeInt, ok = pb.SubCredType_value[strings.ToUpper(vcsType)]
	if !ok || pb.SubCredType(vcsTypeInt) == pb.SubCredType_NIL_SCT {
		ocelog.Log().Fatalf("%s is not a vcs subcredtype, need %s", vcsType, strings.Join(pb.CredType_VCS.SubtypesString(), "|"))
	}
	branchList := strings.Split(branches, ",")
	conf := &changeSetConfig{RemoteConf: rc, AcctRepo: acctRepo, Branches: branchList, Deserializer: deserialize.New(), Producer: nsqpb.GetInitProducer(), OcyValidator: build.GetOcelotValidator()}
	conf.Acct, conf.Repo, err = stringbuilder.GetAcctRepo(acctRepo)
	if err != nil {
		ocelog.Log().Fatal(err)
	}
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
	checker := poll.NewChangeChecker(sig, conf.AcctRepo, conf.VcsType)

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

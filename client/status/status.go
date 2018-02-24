package status

import (
	"bitbucket.org/level11consulting/ocelot/client/commandhelper"
	"github.com/mitchellh/cli"
	"flag"
	"strings"
	"context"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"fmt"
	"bitbucket.org/level11consulting/ocelot/util/cmd_table"
)

const synopsis = "show status of specific acctname/repo, repo or hash"
const help = `
Usage: ocelot status 
	-hash <hash> [optional] if specified, this is the one that takes precendence over other arguments
	-acct-repo <acctname/repo> [optional] if specified, takes precedence over -repo argument
	-repo <repo> [optional] returns status of all repos starting with value
`

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui, config: commandhelper.Config}
	c.init()
	return c
}

type cmd struct {
	UI cli.Ui
	flags   *flag.FlagSet
	config *commandhelper.ClientConfig

	accountRepo string
	repo string
	hash string
}


func (c *cmd) GetClient() models.GuideOcelotClient {
	return c.config.Client
}

func (c *cmd) GetUI() cli.Ui {
	return c.UI
}

func (c *cmd) GetConfig() *commandhelper.ClientConfig {
	return c.config
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return help
}

func (c *cmd) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	//we accept all 3 flags, but prioritize output in the following order: hash, acct-repo, acct
	c.flags.StringVar(&c.hash, "hash", "ERROR", "[optional]  <hash> to display build status")
	c.flags.StringVar(&c.repo, "repo", "ERROR", "[optional]  <repo> to display build status")
	c.flags.StringVar(&c.accountRepo, "acct-repo", "ERROR", "[optional]  <account>/<repo> to display build status")
}

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}

	if c.accountRepo == "ERROR" && c.repo == "ERROR" && c.hash == "ERROR" {
		sha := commandhelper.FindCurrentHash()
		if len(sha) > 0 {
			c.UI.Warn(fmt.Sprintf("no -hash flag passed, using detected hash %s", sha))
			c.hash = sha
		} else {
			c.UI.Error("one of the following flags must be set: -acct-repo, -repo, -hash. see --help")
			return 1
		}
	}

	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}

	//TODO: talk about this with Jessi - what is criteria for project is RUNNING???
	// always respect hash first
	if c.hash != "ERROR" && len(c.hash) > 0 {
		builds, err := c.GetClient().BuildRuntime(ctx, &models.BuildQuery{
			Hash: c.hash,
		})
		if err != nil {
			c.UI.Error(fmt.Sprintf("error retrieving build runtime for hash %s. Error: %s", c.hash, err.Error()))
			return 1
		}

		if len(builds.Builds) > 1 {
			c.UI.Output(cmd_table.SelectFromHashes(builds))
			return 0
		} else if len(builds.Builds) == 0 {
			c.UI.Warn(fmt.Sprintf("no builds found for hash %s", c.hash))
			return 0
		}

		for hash, build := range builds.Builds {
			var possibleErr string
			query := &models.StatusQuery{
				Hash: hash,
			}
			statuses, err := c.GetClient().GetStatus(ctx, query)
			// just because we couldn't get stage details for this hash, doesn't mean it should fail
			if err != nil {
				possibleErr = "\t" + err.Error()
			}
			stagesDetail, color, status := cmd_table.PrintStatusStages(len(build.Ip) > 0, statuses)
			buildStatus := cmd_table.PrintStatusOverview(color, build.AcctName, build.RepoName, hash, status)
			c.UI.Output(buildStatus + possibleErr + stagesDetail)
		}
		return 0
	}

	//respect acct-repo next
	if c.accountRepo != "ERROR" {
		data := strings.Split(c.accountRepo, "/")
		if len(data) != 2  {
			c.UI.Error("flag -acct-repo must be in the format <account>/<repo>. see --help")
			return 1
		}

		query := &models.StatusQuery{
			AcctName: data[0],
			RepoName: data[1],
		}
		statuses, err := c.GetClient().GetStatus(ctx, query)
		if err != nil {
			c.UI.Error(err.Error())
			return 1
		}

		stageStatus, color, status := cmd_table.PrintStatusStages(statuses.BuildSum.BuildDuration < 0, statuses)
		buildStatus := cmd_table.PrintStatusOverview(color, statuses.BuildSum.Account, statuses.BuildSum.Repo, statuses.BuildSum.Hash, status)
		c.UI.Output(buildStatus + stageStatus)
		return 0
	}

	//repo is last
	if c.repo != "ERROR" {
		query := &models.StatusQuery{
			PartialRepo: c.repo,
		}
		statuses, err := c.GetClient().GetStatus(ctx, query)
		if err != nil {
			c.UI.Error(err.Error())
			return 1
		}
		stageStatus, color, status := cmd_table.PrintStatusStages(statuses.BuildSum.BuildDuration < 0, statuses)
		buildStatus := cmd_table.PrintStatusOverview(color, statuses.BuildSum.Account, statuses.BuildSum.Repo, statuses.BuildSum.Hash, status)
		c.UI.Output(buildStatus + stageStatus)
		return 0
	}

	return 0
}
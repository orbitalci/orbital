package status

import (
	"bitbucket.org/level11consulting/ocelot/client/commandhelper"
	"github.com/mitchellh/cli"
	"flag"
	"strings"
	"context"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"fmt"
	"github.com/golang/protobuf/ptypes/wrappers"
	"bitbucket.org/level11consulting/ocelot/util/cmd_table"
)

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

	// always respect hash first
	if c.hash != "ERROR" {
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
		}

		for hash, build := range builds.Builds {
			var status, stagesPrintln string
			var color int

			statuses, err := c.GetClient().StatusByHash(ctx, &wrappers.StringValue{Value: hash})
			// just because we couldn't get stage details for this hash, doesn't mean it should fail
			if err != nil {
				stagesPrintln = "\t" + err.Error()
			}

			if len(build.Ip) > 0 {
				status = "Running"
				color = 33
			} else if statuses != nil {
				if !statuses.BuildSum.Failed {
					status = "PASS"
					color = 32
				} else {
					status = "FAIL"
					color = 31
				}
			}

			if statuses != nil {
				for _, stage := range statuses.Stages {
					stagesPrintln += fmt.Sprintf("\n\t\t\033[0;35m%s\033[0m in %s", stage.Stage, commandhelper.PrettifyTime(stage.StageDuration))
					if statuses.BuildSum.Failed {
						stagesPrintln += fmt.Sprintf("\n\t\t  %s", strings.Join(stage.Messages, "\n\t\t  "))
						if len(stage.Error) > 0 {
							stagesPrintln += fmt.Sprintf(": \033[1;30m%s\033[0m", stage.Error)
						}
					}
				}
			}

			buildStatus := fmt.Sprintf("\033[0;%dm\t %s/%s \t %s \t %s\033[0m", color, build.AcctName, build.RepoName, hash, status)
			c.UI.Output(buildStatus + stagesPrintln)
		}

		return 0
	}

	//respect acct-repo next
	if c.repo != "ERROR" {

	}

	//acct is last
	data := strings.Split(c.accountRepo, "/")
	if len(data) != 2  {
		c.UI.Error("flag -acct-repo must be in the format <account>/<repo>. see --help")
		return 1
	}

	return 0
}

const synopsis = "show status of specific acctname, acctname/repo, or hash"
//TODO: finish writing help once I know how it works
const help = `
Usage: ocelot status -acct-repo <acct>/<repo>
`
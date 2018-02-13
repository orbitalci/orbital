package status

import (
	"bitbucket.org/level11consulting/ocelot/client/commandhelper"
	"github.com/mitchellh/cli"
	"flag"
	"strings"
	"context"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"fmt"
)

const synopsis = "show status of specific acctname, acctname/repo, or hash"
//TODO: finish writing help once I know how it works
const help = `
Usage: ocelot status -acct-repo <acct>/<repo>
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
	account string
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
	c.flags.StringVar(&c.account, "acct", "ERROR", "[optional]  <account> to display build status")
	c.flags.StringVar(&c.accountRepo, "acct-repo", "ERROR", "[optional]  <account>/<repo> to display build status")
}

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}

	if c.accountRepo == "ERROR" && c.account == "ERROR" && c.hash == "ERROR" {
		c.UI.Error("one of the following flags must be set: -acct-repo, -acct, -hash. see --help")
		return 1
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
		//it's okay to iterate here cause list will always contain 1 value
		for hash, build := range builds.Builds {
			var status string
			var color int
			if len(build.Ip) > 0 {
				status = "Running"
				color = 33
			} else {
				status = "Finished "
				color = 34
			}
			buildStatus := fmt.Sprintf("\033[0;%dm\t %s/%s \t %s \t %s\033[0m", color, build.AcctName, build.RepoName, hash, strings.ToUpper(status))
			c.UI.Output(buildStatus)
			return 0
		}
		return 0
	}

	//respect acct-repo next

	//acct is last


	data := strings.Split(c.accountRepo, "/")
	if len(data) != 2  {
		c.UI.Error("flag -acct-repo must be in the format <account>/<repo>. see --help")
		return 1
	}



	return 0
}
package watch

import (
	"bitbucket.org/level11consulting/ocelot/client/commandhelper"
	"fmt"
	"github.com/mitchellh/cli"
	"flag"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"context"
)


const synopsis = "add a repo to ocelot"
const help = `
Usage: ocelot watch -acct-repo <acct>/<repo>
	If an ocelot.yml exists in the root directory of the project, new commits to 
	the project will now trigger builds

` + commandhelper.AcctRepoHelp

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui, config: commandhelper.Config, OcyHelper: &commandhelper.OcyHelper{}}
	c.init()
	return c
}

type cmd struct {
	UI cli.Ui
	flags   *flag.FlagSet
	config *commandhelper.ClientConfig
	*commandhelper.OcyHelper
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
	c.flags.StringVar(&c.OcyHelper.AcctRepo, "acct-repo", "ERROR", "<account>/<repo> to watch")
}


func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	if err := c.OcyHelper.DetectAcctRepo(c.UI); err != nil {
		return 1
	}
	if err := c.OcyHelper.SplitAndSetAcctRepo(c.UI); err != nil {
		return 1
	}

	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}

	_, err := c.config.Client.WatchRepo(ctx, &models.RepoAccount{
		Repo: c.OcyHelper.Repo,
		Account: c.OcyHelper.Account,
	})

	if err != nil {
		c.UI.Error(fmt.Sprintf("unable to watch repo %s/%s! error: %s", c.OcyHelper.Repo, c.OcyHelper.Account, err.Error()))
		return 1
	}

	c.UI.Info(fmt.Sprintf("now watching %s! go on, make a commit and try `ocelot status`", c.OcyHelper.Repo))
	return 0
}

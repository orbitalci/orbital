package polldelete

import (
	"context"
	"flag"
	"fmt"

	"github.com/mitchellh/cli"
	"github.com/shankj3/ocelot/client/commandhelper"
	models "github.com/shankj3/ocelot/models/pb"
)

const synopsis = "delete git poll for repo tracked by ocelot"
const help = `
Usage: ocelot poll delete -acct-repo <acct>/<repo> -cron <cron_string> -branches <branch_list>
	Remove polling from vcs for your repository from ocelot.
For example:
    ocelot poll delete -acct-repo level11consulting/ocelog 
` + commandhelper.AcctRepoHelp

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui, config: commandhelper.Config, OcyHelper: &commandhelper.OcyHelper{}}
	c.init()
	return c
}

type cmd struct {
	UI       cli.Ui
	flags    *flag.FlagSet
	cron     string
	branches string
	config   *commandhelper.ClientConfig
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
	c.SetGitHelperFlags(c.flags, true, false, true)
}

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	if err := c.DetectAcctRepo(c.UI); err != nil {
		commandhelper.Debuggit(c.UI, err.Error())
		return 1
	}
	if err := c.DetectOrConvertVcsType(c.UI); err != nil {
		commandhelper.Debuggit(c.UI, err.Error())
		return 1
	}

	if err := c.OcyHelper.SplitAndSetAcctRepo(c.UI); err != nil {
		return 1
	}
	c.DebugOcyHelper(c.UI)
	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}
	_, err := c.config.Client.DeletePollRepo(ctx, &models.PollRequest{
		Account: c.OcyHelper.Account,
		Repo:    c.OcyHelper.Repo,
	})

	if err != nil {
		c.UI.Error(fmt.Sprintf("unable to delete vcs polling from ocelot for repo %s/%s! error: %s", c.OcyHelper.Repo, c.OcyHelper.Account, err.Error()))
		return 1
	}

	c.UI.Info(fmt.Sprintf("Deleted polling for %s!", c.OcyHelper.AcctRepo))
	return 0
}

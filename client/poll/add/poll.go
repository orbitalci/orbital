package polladd

import (
	"context"
	"flag"
	"fmt"
	"github.com/shankj3/ocelot/client/commandhelper"
	models "github.com/shankj3/ocelot/models/pb"
	"strings"

	"github.com/gorhill/cronexpr"
	"github.com/mitchellh/cli"
)

const synopsis = "set up repo to be polled by ocelot"
const help = `
Usage: ocelot poll -acct-repo <acct>/<repo> -cron <cron_string> -branches <branch_list>
	Set up polling from vcs for your repository. Will run on the interval that you set in the cron string.
For example:
    ocelot poll -acct-repo level11consulting/ocelog -cron "5 4 * * *" -branches master,test
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
	c.flags.StringVar(&c.AcctRepo, "acct-repo", "ERROR", "<account>/<repo> to watch")
	c.flags.StringVar(&c.cron, "cron", "ERROR", "cron string for polling repo ")
	c.flags.StringVar(&c.branches, "branches", "ERROR", "comma separated list of branches to poll vcs for")
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
	if c.cron == "ERROR" {
		commandhelper.Debuggit(c.UI, strings.Join(args[:], ","))
		c.UI.Error("-cron is a required flag")
		return 1
	}
	if c.branches == "ERROR" {
		c.UI.Error("-branches is a required flag")
		return 1
	}

	if _, err := cronexpr.Parse(c.cron); err != nil {
		errStr := `Supplied cron expression is not valid! Received: %s
Error: %s`
		c.UI.Error(fmt.Sprintf(errStr, c.cron, err.Error()))
		return 1
	}

	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}
	pr := &models.PollRequest{
		Account:  c.OcyHelper.Account,
		Repo:     c.OcyHelper.Repo,
		Cron:     c.cron,
		Branches: c.branches,
		Type:     c.VcsType,
	}
	commandhelper.Debuggit(c.UI, fmt.Sprintf("%#v", pr))
	_, err := c.config.Client.PollRepo(ctx, pr)
	if err != nil {
		c.UI.Error(fmt.Sprintf("unable to set up vcs polling for repo %s/%s! error: %s", c.OcyHelper.Repo, c.OcyHelper.Account, err.Error()))
		return 1
	}

	c.UI.Info(fmt.Sprintf("Now set up polling for %s! Will check vcs on the cron interval %s, and if there are changes for any of the specified branches, it will trigger a build.", c.OcyHelper.AcctRepo, c.cron))
	return 0
}

package watch

import (
	"bitbucket.org/level11consulting/ocelot/client/commandhelper"
	"strings"
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
	c.flags.StringVar(&c.accountRepo, "acct-repo", "ERROR", "<account>/<repo> to watch")
}


func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	if c.accountRepo == "ERROR" {
		acctRepo, err := commandhelper.FindAcctRepo()
		// should we even be reporting this error to the user? that git command failed?
		if err != nil {
			commandhelper.Debuggit(c, "error!!! " + err.Error())
			c.UI.Error("flag -acct-repo must be in the format <account>/<repo> or you must be in the directory you wish to view a summary of. see --help")
			return 1
		}
		c.accountRepo = acctRepo
	}

	data := strings.Split(c.accountRepo, "/")
	if len(data) != 2  {
		c.UI.Error("flag -acct-repo must be in the format <account>/<repo>. see --help")
		return 1
	}
	account := data[0]
	repo := data[1]

	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}

	_, err := c.config.Client.WatchRepo(ctx, &models.RepoAccount{
		Repo: repo,
		Account: account,
	})

	if err != nil {
		c.UI.Error(fmt.Sprintf("unable to watch repo %s/%s! error: %s", repo, account, err.Error()))
		return 1
	}

	c.UI.Info(fmt.Sprintf("now watching %s! go on, make a commit and try `ocelot status`", repo))
	return 0
}

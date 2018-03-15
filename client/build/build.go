package build

import (
	"bitbucket.org/level11consulting/ocelot/client/commandhelper"
	"fmt"
	"github.com/mitchellh/cli"
	"flag"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"context"
)


const synopsis = "build a project hash"
const help = `
Usage: ocelot build -acct-repo <acct>/<repo> -hash <git_hash>
	Triggers a build to happen for the specified account/repository and hash starting with <git_hash>
	If you running this command from the directory of your cloned git repo, ocelot will run some git commands
	to detect the account and repo name from the origin url, and it will detect the last pushed commit. 
	Those values will be used to trigger the build.
`

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
	//TODO: trigger also by build id? Need to standardize across commands
	c.flags.StringVar(&c.OcyHelper.AcctRepo, "acct-repo", "ERROR", "<account>/<repo> to build")
	c.flags.StringVar(&c.OcyHelper.Hash, "hash", "ERROR", "hash to build")
}


func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		commandhelper.Debuggit(c, err.Error())
		return 1
	}

	if err := c.OcyHelper.DetectHash(c); err != nil {
		commandhelper.Debuggit(c, err.Error())
		return 1
	}

	if err := c.OcyHelper.DetectAcctRepo(c); err != nil {
		commandhelper.Debuggit(c, err.Error())
		return 1
	}
	if err := c.OcyHelper.SplitAndSetAcctRepo(c); err != nil {
		commandhelper.Debuggit(c, err.Error())
	}

	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}

	_, err := c.config.Client.BuildRepoAndHash(ctx, &models.AcctRepoAndHash{
		AcctRepo: c.OcyHelper.AcctRepo,
		PartialHash: c.OcyHelper.Hash,
	})

	if err != nil {
		c.UI.Error(fmt.Sprintf("unable to build repo %s! error: %s", c.OcyHelper.AcctRepo, err.Error()))
		return 1
	}

	c.UI.Info(fmt.Sprintf("triggered build for %s", c.OcyHelper.AcctRepo))
	return 0
}

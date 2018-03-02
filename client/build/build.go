package build

import (
	"bitbucket.org/level11consulting/ocelot/client/commandhelper"
	"strings"
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
	c.flags.StringVar(&c.accountRepo, "acct-repo", "ERROR", "<account>/<repo> to build")
	c.flags.StringVar(&c.hash, "hash", "ERROR", "hash to build")
}


func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}

	if c.hash == "ERROR" {
		sha := commandhelper.FindCurrentHash()
		if len(sha) > 0 {
			c.UI.Warn(fmt.Sprintf("no -hash flag passed, using detected hash %s", sha))
			c.hash = sha
		}
	}

	if c.accountRepo == "ERROR" || c.hash == "ERROR" {
		c.UI.Error("flag -acct-repo must be in the format <account>/<repo> and -hash must be the start of a valid hash")
		return 1
	}

	data := strings.Split(c.accountRepo, "/")
	if len(data) != 2  {
		c.UI.Error("flag -acct-repo must be in the format <account>/<repo>. see --help")
		return 1
	}

	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}

	_, err := c.config.Client.BuildRepoAndHash(ctx, &models.AcctRepoAndHash{
		AcctRepo: c.accountRepo,
		PartialHash: c.hash,
	})

	if err != nil {
		c.UI.Error(fmt.Sprintf("unable to build repo %s! error: %s", c.accountRepo, err.Error()))
		return 1
	}

	c.UI.Info(fmt.Sprintf("triggered build for %s", c.accountRepo))
	return 0
}

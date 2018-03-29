package build

import (
	"bitbucket.org/level11consulting/ocelot/client/commandhelper"
	"github.com/mitchellh/cli"
	"flag"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"context"
	"io"
)


const synopsis = "build a project hash"
const help = `
Usage: ocelot build -acct-repo <acct>/<repo> -hash <git_hash> -branch <branch>
	Triggers a build to happen for the specified account/repository and hash starting with <git_hash>
	If you running this command from the directory of your cloned git repo, ocelot will run some git commands
	to detect the account and repo name from the origin url, and it will detect the last pushed commit. 
	Those values will be used to trigger the build. If passing in a hash that hasn't been built before, 
	branch MUST BE SPECIFIED. If a build corresponding with the passed hash (or starts with passed hash value) 
	already exists, then the branch can be inferred and you will not be required to pass a branch flag. 
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
	Branch 	string
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
	c.flags.StringVar(&c.AcctRepo, "acct-repo", "ERROR", "<account>/<repo> to build")
	c.flags.StringVar(&c.Hash, "hash", "ERROR", "hash to build")
	c.flags.StringVar(&c.Branch, "branch", "ERROR", "branch to build (only required if passing a previously un-built hash or overriding the branch associated with a previous build)")
}


func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		commandhelper.Debuggit(c.UI, err.Error())
		return 1
	}

	if err := c.DetectHash(c.UI); err != nil {
		commandhelper.Debuggit(c.UI, err.Error())
		return 1
	}

	if err := c.DetectAcctRepo(c.UI); err != nil {
		commandhelper.Debuggit(c.UI, err.Error())
		return 1
	}
	if err := c.SplitAndSetAcctRepo(c.UI); err != nil {
		commandhelper.Debuggit(c.UI, err.Error())
	}

	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}

	buildRequest := &models.BuildReq{
		AcctRepo: c.AcctRepo,
		Hash: c.Hash,
	}

	if c.Branch != "ERROR" && len(c.Branch) > 0 {
		buildRequest.Branch = c.Branch
	}

	stream, err := c.config.Client.BuildRepoAndHash(ctx, buildRequest)

	if err != nil {
		commandhelper.UIErrFromGrpc(err, c.UI, "Unable to get build results from admin")
		return 1
	}

	err = c.HandleStreaming(c.UI, stream)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	return 0
}

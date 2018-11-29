package build

import (
	"context"
	"flag"
	"fmt"

	"github.com/mitchellh/cli"
	help "github.com/shankj3/ocelot/client/commandhelper"
	models "github.com/shankj3/ocelot/models/pb"

)

const synopsis = "build a specific commit of a repository"
const helpmsg = `
Usage: ocelot build -acct-repo <acct>/<repo> -hash <git_hash> -branch <branch> [-latest]

	Triggers a build to happen for the specified account/repository and hash starting with <git_hash>
	If you running this command from the directory of your cloned git repo, ocelot will run some git commands
	to detect the account and repo name from the origin url, and it will detect the last pushed commit. 
	Those values will be used to trigger the build. If passing in a hash that hasn't been built before, 
	branch MUST BE SPECIFIED. If a build corresponding with the passed hash (or starts with passed hash value) 
	already exists, then the branch can be inferred and you will not be required to pass a branch flag. If you 
	wish to run the remote tip of a branch, pass in the -latest flag

Example: 
	1. Building a commit that has already been built before, while in the directory of the git repo:
	$ ocelot build 
	In this case, the client will detect the vcs account, repository, and the hash by running git commands. Since it
	has been built previously, the branch will already be known to ocelot.

	2. Building a commit that you just pushed to master, while in the directory of the git repo:
	$ ocelot build -branch master
	Here, a new commit is being detected by the ocelot client, and therefore it will not be able to guess which branch it belongs to. 
	
	3. Building the tip of a remote branch (let's say, feature2)
	$ ocelot build -branch feature2 -latest
	The -latest flag will tell ocelot to search for the head commit at the remote git repository (branch feature2) and build using that hash. 
`

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui, config: help.Config, OcyHelper: &help.OcyHelper{}}
	c.init()
	return c
}

type cmd struct {
	UI     cli.Ui
	flags  *flag.FlagSet
	config *help.ClientConfig
	Branch string
	vcstyp string
	force  bool
	latest bool
	*help.OcyHelper
}

func (c *cmd) GetClient() models.GuideOcelotClient {
	return c.config.Client
}

func (c *cmd) GetUI() cli.Ui {
	return c.UI
}

func (c *cmd) GetConfig() *help.ClientConfig {
	return c.config
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return helpmsg
}

func (c *cmd) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	//TODO: trigger also by build id? Need to standardize across commands
	c.flags.StringVar(&c.Branch, "branch", "ERROR", "branch to build (only required if passing a previously un-built hash or overriding the branch associated with a previous build)")
	c.flags.BoolVar(&c.force, "force", false, "force the build to be queued even if it is not one of the accepted branches")
	c.flags.BoolVar(&c.latest, "latest", false, "use -latest to find the latest commit of the acct/repo at the branch denoted by -branch")
	c.SetGitHelperFlags(c.flags, true, true, true)
}

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		help.Debuggit(c.UI, err.Error())
		return 1
	}
	if c.latest {
		c.Hash = ""
	} else {
		if err := c.DetectHash(c.UI); err != nil {
			help.Debuggit(c.UI, err.Error())
			return 1
		}
	}

	if err := c.DetectAcctRepo(c.UI); err != nil {
		help.Debuggit(c.UI, err.Error())
		return 1
	}
	if err := c.DetectOrConvertVcsType(c.UI); err != nil {
		help.Debuggit(c.UI, err.Error())
		// if we can't set the vcs type rightn ow, that's alright. admin is going to try to figure out who owns this anyway
		//return 1
	}

	if err := c.SplitAndSetAcctRepo(c.UI); err != nil {
		help.Debuggit(c.UI, err.Error())
	}
	c.DebugOcyHelper(c.UI)
	ctx := context.Background()
	if err := help.CheckConnection(c, ctx); err != nil {
		return 1
	}

	buildRequest := &models.BuildReq{
		AcctRepo: c.AcctRepo,
		Hash:     c.Hash,
		Force:    c.force,
		VcsType:  c.VcsType,
	}

	if c.Branch != "ERROR" && len(c.Branch) > 0 {
		buildRequest.Branch = c.Branch
	}
	help.Debuggit(c.UI, fmt.Sprintf("%#v", buildRequest))
	stream, err := c.config.Client.BuildRepoAndHash(ctx, buildRequest)

	if err != nil {
		help.UIErrFromGrpc(err, c.UI, "Unable to get build results from admin")
		return 1
	}

	err = c.HandleStreaming(c.UI, stream)
	if err != nil {
		help.Debuggit(c.UI, err.Error())
		return 1
	}

	return 0
}

package kill

import (
	"github.com/mitchellh/cli"
	"github.com/shankj3/ocelot/client/commandhelper"
	bld "github.com/shankj3/ocelot/common/build"
	models "github.com/shankj3/ocelot/models/pb"

	"context"
	"flag"
	"fmt"
)

const synopsis = "kill an active build for a hash"
const help = `
Usage: ocelot kill 
	- hash <hash> [required] a partial or full hash of the build you'd like to kill'
`

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui, config: commandhelper.Config, OcyHelper: &commandhelper.OcyHelper{}}
	c.init()
	return c
}

type cmd struct {
	UI     cli.Ui
	flags  *flag.FlagSet
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

//NOTE: this assumes that only one build is happening with this hash!!!!!
func (c *cmd) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.StringVar(&c.OcyHelper.Hash, "hash", "ERROR", "[REQUIRED] <hash> to kill")
}

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}
	var build *models.BuildRuntimeInfo
	var err error

	if c.Hash == "ERROR" {
		if err := c.OcyHelper.DetectHash(c.UI); err != nil {
			commandhelper.Debuggit(c.UI, err.Error())
			return 1
		}
	}

	build, err = c.config.Client.FindWerker(ctx, &models.BuildReq{
		Hash: c.Hash,
	})

	if err != nil {
		c.UI.Error(fmt.Sprintf("error looking up build for hash %s: %s", c.Hash, err.Error()))
		return 1
	}

	client, err := bld.CreateBuildClient(build)
	commandhelper.Debuggit(c.UI, fmt.Sprintf("dialing werker at %s:%s", build.GetIp(), build.GetGrpcPort()))
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error dialing the werker at %s:%s! Error: %s", build.GetIp(), build.GetGrpcPort(), err.Error()))
		return 1
	}

	stream, err := client.KillHash(ctx, &models.Request{Hash: build.Hash})
	if err != nil {
		commandhelper.UIErrFromGrpc(err, c.UI, fmt.Sprintf("Unable to get build info stream from client at %s:%s!", build.GetIp(), build.GetGrpcPort()))
		return 1
	}

	err = c.HandleStreaming(c.UI, stream, c.config.Theme.NoColor)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	return 0
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return help
}

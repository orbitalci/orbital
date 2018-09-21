package output

import (
	"context"
	"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"github.com/shankj3/ocelot/client/commandhelper"
	bldr "github.com/shankj3/ocelot/common/build"
	models "github.com/shankj3/ocelot/models/pb"
	"io"
)

const synopsis = "stream logs on running or completed build"
const help = `
Usage: ocelot logs --hash <git_hash>
  Will stream logs of a running or completed build identified by the hash. 
  If the build is: 
    - completed:   logs will stream from storage via the admin
    - in progress: logs will stream from the werker node running the build 
    - not found:   he cli will return an error and exit. 
  If there are multiple builds of the same hash, it will stream the latest found in storage.
  If you have the build id from running 'ocelot summary', you can also stream old logs by running: 
    ocelot logs -build-id <id>
`

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui, config: commandhelper.Config, OcyHelper: &commandhelper.OcyHelper{}}
	c.init()
	return c
}

type cmd struct {
	UI      cli.Ui
	flags   *flag.FlagSet
	config  *commandhelper.ClientConfig
	buildId int
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

func (c *cmd) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.StringVar(&c.OcyHelper.Hash, "hash", "ERROR",
		"*REQUIRED* hash to get build data of")
	c.flags.IntVar(&c.buildId, "build-id", 0, "build id from build_summary table, if you do not want to get the last")
}

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}
	var build *models.Builds
	var err error

	if c.buildId != 0 {
		build, err = c.config.Client.BuildRuntime(ctx, &models.BuildQuery{BuildId: int64(c.buildId)})
	} else {
		if err := c.OcyHelper.DetectHash(c.UI); err != nil {
			commandhelper.Debuggit(c.UI, err.Error())
			return 1
		}
		build, err = c.config.Client.BuildRuntime(ctx, &models.BuildQuery{Hash: c.OcyHelper.Hash})
	}

	if err != nil {
		c.UI.Error("unable to get build runtime! error: " + err.Error())
		return 1
	}

	if len(build.Builds) > 1 {
		c.UI.Output(commandhelper.SelectFromHashes(build, c.config.Theme))
		return 0
	} else if len(build.Builds) == 1 {
		for _, build := range build.Builds {
			if build.Done {
				commandhelper.Debuggit(c.UI, "streaming from storage")
				return c.fromStorage(ctx, build.Hash, int64(c.buildId))
			} else {
				commandhelper.Debuggit(c.UI, "streaming from werker")
				return c.fromWerker(ctx, build)
			}
		}
	} else {
		c.UI.Warn(fmt.Sprintf("Warning: No builds found for entry: %s", c.OcyHelper.Hash))
		return 0
	}
	return 0
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return help
}

func (c *cmd) fromStorage(ctx context.Context, hash string, id int64) int {
	var stream models.GuideOcelot_LogsClient
	var err error
	stream, err = c.config.Client.Logs(ctx, &models.BuildQuery{BuildId: id, Hash: hash})
	if err != nil {
		commandhelper.UIErrFromGrpc(err, c.UI, "Unable to get stream from admin.")
		return 1
	}
	for {
		line, err := stream.Recv()
		if err == io.EOF {
			stream.CloseSend()
			return 0
		} else if err != nil {
			//todo: implement this
			c.UI.Warn("Warning: partial matching is not yet supported for streaming from storage")
			commandhelper.UIErrFromGrpc(err, c.UI, "Error streaming from storage via admin.")
			return 1
		}
		c.UI.Info(line.GetOutputLine())
	}
}

func (c *cmd) fromWerker(ctx context.Context, build *models.BuildRuntimeInfo) int {
	client, err := bldr.CreateBuildClient(build)
	commandhelper.Debuggit(c.UI, fmt.Sprintf("dialing werker at %s:%s", build.GetIp(), build.GetGrpcPort()))
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error dialing the werker at %s:%s! Error: %s", build.GetIp(), build.GetGrpcPort(), err.Error()))
		return 1
	}

	stream, err := client.BuildInfo(ctx, &models.Request{Hash: build.GetHash()})
	if err != nil {
		commandhelper.UIErrFromGrpc(err, c.UI, fmt.Sprintf("Unable to get build info stream from client at %s:%s!", build.GetIp(), build.GetGrpcPort()))
		return 1
	}

	err = c.HandleStreaming(c.UI, stream)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	return 0
}

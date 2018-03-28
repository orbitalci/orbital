package output

import (
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/client/commandhelper"
	pb "bitbucket.org/level11consulting/ocelot/werker/protobuf"
	"context"
	"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"google.golang.org/grpc"
	"io"
	"os"
	"os/signal"
	"syscall"
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
	UI cli.Ui
	flags   *flag.FlagSet
	config *commandhelper.ClientConfig
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
		if err := c.OcyHelper.DetectHash(c); err != nil {
			commandhelper.Debuggit(c, err.Error())
			return 1
		}
		build, err = c.config.Client.BuildRuntime(ctx, &models.BuildQuery{Hash: c.OcyHelper.Hash})
	}

	if err != nil {
		c.UI.Error("unable to get build runtime! error: " + err.Error())
		return 1
	}

	if len(build.Builds) > 1 {
		c.UI.Output(commandhelper.SelectFromHashes(build))
		return 0
	} else if len(build.Builds) == 1 {
		for _, build := range build.Builds {
			if build.Done {
				commandhelper.Debuggit(c, "streaming from storage")
				return c.fromStorage(ctx, build.Hash)
			} else {
				commandhelper.Debuggit(c, "streaming from werker")
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


func (c *cmd) fromStorage(ctx context.Context, hash string) int {
	stream, err := c.config.Client.Logs(ctx, &models.BuildQuery{Hash: hash})
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

func (c *cmd) fromWerker(ctx context.Context, build models.BuildRuntime) int {
	var opts []grpc.DialOption
	// right now werker is insecure
	opts = append(opts, grpc.WithInsecure())
	client, err := build.CreateBuildClient(opts)
	commandhelper.Debuggit(c, fmt.Sprintf("dialing werker at %s:%s", build.GetIp(), build.GetGrpcPort()))
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error dialing the werker at %s:%s! Error: %s", build.GetIp(), build.GetGrpcPort(), err.Error()))
		return 1
	}

	stream, err := client.BuildInfo(ctx, &pb.Request{Hash: build.GetHash()})
	if err != nil {
		commandhelper.UIErrFromGrpc(err, c.UI, fmt.Sprintf("Unable to get build info stream from client at %s:%s!", build.GetIp(), build.GetGrpcPort()))
		return 1
	}
	interrupt := make(chan os.Signal, 2)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-interrupt
		c.UI.Info("received ctl-c, exiting")
		stream.CloseSend()
		os.Exit(1)
	}()
	for {
		commandhelper.Debuggit(c, "receiving stream")
		line, err := stream.Recv()
		if err == io.EOF {
			stream.CloseSend()
			return 0
		} else if err != nil {
			commandhelper.UIErrFromGrpc(err, c.UI, "Error streaming from werker.")
			return 1
		}
		c.UI.Info(line.GetOutputLine())
	}
}



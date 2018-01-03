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
)

const synopsis = "stream logs on running or completed build"
const help = `
Usage: ocelot logs --hash <git_hash>
  Will stream logs of a running or completed build identified by the hash. If the build is still running, it will stream from the werker.
  If it is finished, will stream from storage using the admin as an intermediary.
  An error will be returned and the tool will exit if that hash is not found.
  If there are multiple builds of the same hash, it will stream the latest.
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


func (c *cmd) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.StringVar(&c.hash, "hash", "ERROR",
		"*REQUIRED* hash to get build data of")
}

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	if c.hash == "ERROR" {
		c.UI.Error("flag --hash is required, otherwise there is no build to tail")
		return 1
	}
	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}
	build, err := c.config.Client.BuildRuntime(ctx, &models.BuildQuery{Hash: c.hash})
	if err != nil {
		c.UI.Error("unable to get build runtime! error: " + err.Error())
		return 1
	}
	if build.Done {
		return c.fromStorage(ctx)
	} else {
		return c.fromWerker(ctx, build)
	}

	return 0
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return help
}


func (c *cmd) fromStorage(ctx context.Context) int {
	stream, err := c.config.Client.Logs(ctx, &models.BuildQuery{Hash: c.hash})
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
			commandhelper.UIErrFromGrpc(err, c.UI, "Error streaming from storage via admin.")
			return 1
		}
		c.UI.Info(line.GetOutputLine())
	}
}

func (c *cmd) fromWerker(ctx context.Context, build *models.BuildRuntimeInfo) int {
	var opts []grpc.DialOption
	// right now werker is insecure
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(build.Ip + ":" + build.GrpcPort, opts...)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error dialing the werker at %s:%s! Error: %s", build.Ip, build.GrpcPort, err.Error()))
		return 1
	}
	client := pb.NewBuildClient(conn)
	stream, err := client.BuildInfo(ctx, &pb.Request{Hash: c.hash})
	if err != nil {
		commandhelper.UIErrFromGrpc(err, c.UI, fmt.Sprintf("Unable to get build info stream from client at %s:%s!", build.Ip, build.GrpcPort))
		return 1
	}
	for {
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



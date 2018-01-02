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

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui, config: commandhelper.Config}
	c.init()
	return c
}

type cmd struct {
	UI cli.Ui
	flags   *flag.FlagSet
	streamCli pb.Build_BuildInfoClient
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
		c.UI.Error("stream from storage not yet implemented")
		// todo: implement c.config.Client.Logs() then call it here
		return 1
	} else {
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
			c.UI.Error(fmt.Sprintf("Unable to get build info stream from client at %s:%s! Error: %s", build.Ip, build.GrpcPort, err.Error()))
		}
		for {
			line, err := stream.Recv()
			if err == io.EOF {
				stream.CloseSend()
				return 0
			} else if err != nil {
				c.UI.Error(fmt.Sprintf("Error streaming: %s", err.Error()))
				return 1
			}
			c.UI.Info(line.GetOutputLine())
		}
	}

	return 0
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return help
}


const synopsis = "list all credentials added to ocelot"
const help = `
Usage: ocelot creds list

  Will list all credentials that have been added to ocelot. //todo filter on acct name
`

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
		"hash to get build data of")
}

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial("localhost:9099", opts...)
	if err != nil {
		panic("err!" + err.Error())
	}
	client := pb.NewBuildClient(conn)
	stream, _ := client.BuildInfo(ctx, &pb.Request{Hash: c.hash})
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
	//msg, err := c.config.Client.GetAllCreds(ctx, &protoReq)
	//if err != nil {
	//	c.UI.Error(fmt.Sprint("Could not get list of credentials!\n Error: ", err.Error()))
	//	return 1
	//}
	//if len(msg.RepoCreds.Repo) > 0 {
	//	repocredslist.Header(c.UI)
	//	for _, oneline := range msg.RepoCreds.Repo {
	//		c.UI.Info(repocredslist.Prettify(oneline))
	//	}
	//} else {
	//	repocredslist.NoDataHeader(c.UI)
	//}
	//
	//if len(msg.VcsCreds.Vcs) > 0 {
	//	buildcredslist.Header(c.UI)
	//	for _, oneline :=  range msg.VcsCreds.Vcs {
	//		c.UI.Info(buildcredslist.Prettify(oneline))
	//	}
	//} else {
	//	buildcredslist.NoDataHeader(c.UI)
	//}

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

package buildcredslist

import (
	"bitbucket.org/level11consulting/ocelot/admin"
	"context"
	"flag"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mitchellh/cli"
)

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui}
	c.init()
	return c
}

type cmd struct {
	UI cli.Ui
	flags   *flag.FlagSet
}

func (c *cmd) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
}

func (c *cmd) Run(args []string) int {
	client, err := admin.GetClient("localhost:10000")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	var protoReq empty.Empty
	msg, err := client.GetCreds(ctx, &protoReq)
	if err != nil {
		c.UI.Error(fmt.Sprint("Could not get list of credentials!\n Error: ", err.Error()))
	}
	pretty := `ClientId: %s
ClientSecret: %s
TokenURL: %s
AcctName: %s
Type: %s

`
	for _, oneline := range msg.Credentials {
		c.UI.Info(fmt.Sprintf(pretty, oneline.ClientId, oneline.ClientSecret, oneline.TokenURL, oneline.AcctName, oneline.Type))
	}
	return 0
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return help
}

const synopsis = "List all credentials used for tracking repositories to build"
const help = `
Usage: ocelot creds list

  Retrieves all credentials that ocelot uses to track repositories
`

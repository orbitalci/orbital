package buildcredslist

import (
	"bitbucket.org/level11consulting/ocelot/admin"
	"bitbucket.org/level11consulting/ocelot/admin/models"
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
	client models.GuideOcelotClient
}

func (c *cmd) init() {
	var err error
	//todo: THIS IS HARDCODED! BAD!
	c.client, err = admin.GetClient("localhost:10000")
	if err != nil {
		panic(err)
	}
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
}

func (c *cmd) Run(args []string) int {
	ctx := context.Background()
	var protoReq empty.Empty
	msg, err := c.client.GetCreds(ctx, &protoReq)
	if err != nil {
		c.UI.Error(fmt.Sprint("Could not get list of credentials!\n Error: ", err.Error()))
	}
	for _, oneline := range msg.Credentials {
		c.UI.Info(prettify(oneline))
	}
	return 0
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return help
}

func prettify(cred *models.Credentials) string {
	pretty := `ClientId: %s
ClientSecret: %s
TokenURL: %s
AcctName: %s
Type: %s

`
	return fmt.Sprintf(pretty, cred.ClientId, cred.ClientSecret, cred.TokenURL, cred.AcctName, cred.Type)
}


const synopsis = "List all credentials used for tracking repositories to build"
const help = `
Usage: ocelot creds list

  Retrieves all credentials that ocelot uses to track repositories
`

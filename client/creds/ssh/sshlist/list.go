package sshlist

import (
	"context"
	"flag"
	"fmt"
	"github.com/shankj3/ocelot/client/commandhelper"
	models "github.com/shankj3/ocelot/models/pb"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mitchellh/cli"
)

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui, config: commandhelper.Config}
	c.init()
	return c
}

type cmd struct {
	UI            cli.Ui
	flags         *flag.FlagSet
	config        *commandhelper.ClientConfig
	accountFilter string
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
}

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}
	var protoReq empty.Empty
	msg, err := c.config.Client.GetSSHCreds(ctx, &protoReq)
	if err != nil {
		c.UI.Error(fmt.Sprint("Could not get list of credentials!\n Error: ", err.Error()))
		return 1
	}
	printed := false
	Header(c.UI)
	for _, oneline := range msg.Keys {
		if c.accountFilter == "" || oneline.AcctName == c.accountFilter {
			c.UI.Output(Prettify(oneline))
			printed = true
		}
	}
	if printed == false {
		NoDataHeader(c.UI)
	}
	return 0
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return help
}

func Header(ui cli.Ui) {
	ui.Output("\n--- SSH Key Credentials ---\n")
}

func NoDataHeader(ui cli.Ui) {
	ui.Warn("\n--- No SSH Key Credentials Found ---\n")
}

func Prettify(cred *models.SSHKeyWrapper) string {
	pretty := `Acccount: %s
Identifier: %s

`
	return fmt.Sprintf(pretty, cred.AcctName, cred.Identifier)
}

const synopsis = "List all credentials used for tracking repositories to build"
const help = `
Usage: ocelot creds ssh list

  Retrieves all ssh keys that ocelot has for use in builds. 
`

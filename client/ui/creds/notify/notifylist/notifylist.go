package notifylist

import (
	"context"
	"flag"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mitchellh/cli"
	"github.com/level11consulting/orbitalci/client/commandhelper"
	models "github.com/level11consulting/orbitalci/models/pb"
)

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui, config: commandhelper.Config}
	c.init()
	return c
}

type cmd struct {
	UI            cli.Ui
	flags         *flag.FlagSet
	accountFilter string
	config        *commandhelper.ClientConfig
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
	c.flags.StringVar(&c.accountFilter, "account", "",
		"account name to filter on")
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
	msg, err := c.config.Client.GetNotifyCreds(ctx, &protoReq)
	if err != nil {
		c.UI.Error(fmt.Sprint("Could not get list of credentials!\n Error: ", err.Error()))
		return 1
	}
	printed := false
	Header(c.UI)
	for _, oneline := range msg.Creds {
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
	ui.Output("\n--- Notify Credentials ---\n")
}

func NoDataHeader(ui cli.Ui) {
	ui.Warn("\n--- No Notify Credentials Found! ---")
}

func Prettify(cred *models.NotifyCreds) string {
	if cred.DetailUrlBase == "" {
		str := `AcctName: %s
Identifier: %s
Type: %s
`
		return fmt.Sprintf(str, cred.AcctName, cred.Identifier, cred.SubType)
	} else {
		str := `AcctName: %s
Identifier: %s
Type: %s
Detail Url: %s
`
		return fmt.Sprintf(str, cred.AcctName, cred.Identifier, cred.SubType, cred.DetailUrlBase)
	}
}

const synopsis = "List all credentials associated with notification integration"
const help = `
Usage: ocelot creds notify list

  Retrieves all credentials that ocelot uses to notify on status of builds
`

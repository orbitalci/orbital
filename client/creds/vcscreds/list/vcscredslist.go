package buildcredslist

import (
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/client/commandhelper"
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mitchellh/cli"
)

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui, config: commandhelper.Config}
	c.init()
	return c
}

type cmd struct {
	UI cli.Ui
	flags   *flag.FlagSet
	config *commandhelper.ClientConfig
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
	msg, err := c.config.Client.GetVCSCreds(ctx, &protoReq)
	if err != nil {
		c.UI.Error(fmt.Sprint("Could not get list of credentials!\n Error: ", err.Error()))
	}
	printed := false
	Header(c.UI)
	for _, oneline := range msg.Vcs {
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
	ui.Output("\n--- Admin Credentials ---\n")
}

func NoDataHeader(ui cli.Ui) {
	ui.Warn("\n--- No Admin Credentials Found ---\n")
}


func Prettify(cred *models.VCSCreds) string {
	pretty := `ClientId: %s
ClientSecret: %s
TokenURL: %s
AcctName: %s
SubType: %s
Identifier: %s
[%s]

`
	return fmt.Sprintf(pretty, cred.ClientId, cred.ClientSecret, cred.TokenURL, cred.AcctName, strings.ToLower(cred.SubType.String()), cred.Identifier, cred.SshFileLoc)
}


const synopsis = "List all credentials used for tracking repositories to build"
const help = `
Usage: ocelot creds vcs list

  Retrieves all credentials that ocelot uses to track repositories
`

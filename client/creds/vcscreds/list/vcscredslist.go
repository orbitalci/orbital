package buildcredslist

import (
	"context"
	"flag"
	"fmt"
	"github.com/shankj3/ocelot/client/commandhelper"
	models "github.com/shankj3/ocelot/models/pb"
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
		return 1
	}
	printed := false
	Header(c.UI)
	for _, oneline := range msg.Vcs {
		if c.accountFilter == "" || oneline.AcctName == c.accountFilter {
			//todo: do the commented out thing
			//commandhelper.Debuggit(c.UI, oneline.SshFileLoc)
			//if oneline.SshFileLoc == "" {
			//	oneline.SshFileLoc = c.config.Theme.Warning.Sprint("No SSH Key")
			//} else {
			//	oneline.SshFileLoc = c.config.Theme.Passed.Sprint("SSH Key on File")
			//}
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

func prettifyGithub(cred *models.VCSCreds) string {
	pretty := `AcctName: %s
ClientSecret: %s
SubType: %s
Identifier: %s
[%s]`
	return fmt.Sprintf(pretty, cred.AcctName, cred.ClientSecret, cred.SubType, cred.Identifier, cred.SshFileLoc)
}

func prettifyBitbucket(cred *models.VCSCreds) string {
	pretty := `AcctName: %s
ClientId: %s
ClientSecret: %s
TokenURL: %s
SubType: %s
Identifier: %s
[%s]

`
	return fmt.Sprintf(pretty, cred.AcctName,  cred.ClientId, cred.ClientSecret, cred.TokenURL, strings.ToLower(cred.SubType.String()), cred.Identifier, cred.SshFileLoc)
}

func Prettify(cred *models.VCSCreds) string {
	switch cred.SubType {
	case models.SubCredType_GITHUB:
		return prettifyGithub(cred)
	case models.SubCredType_BITBUCKET:
		return prettifyBitbucket(cred)
	default:
		return fmt.Sprintf("ERROR! UNKNOWN TYPE OF VCS CRED: %s", cred.SubType.String())
	}
}

const synopsis = "List all credentials used for tracking repositories to build"
const help = `
Usage: ocelot creds vcs list

  Retrieves all credentials that ocelot uses to track repositories
`

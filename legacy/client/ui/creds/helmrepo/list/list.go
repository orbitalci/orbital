package helmrepolist

import (
	"bytes"
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
	msg, err := c.config.Client.GetGenericCreds(ctx, &protoReq)
	if err != nil {
		c.UI.Error(fmt.Sprint("Could not get list of credentials!\n Error: ", err.Error()))
		return 1
	}
	organized := organize(msg)
	var creds2list = make(map[string]*[]*models.GenericCreds)
	if c.accountFilter != "" {
		if filtered := organized[c.accountFilter]; filtered != nil {
			creds2list[c.accountFilter] = organized[c.accountFilter]
		}
	} else {
		creds2list = organized
	}
	if len(creds2list) == 0 {
		NoDataHeader(c.UI)
		return 0
	} else {
		Header(c.UI)
	}
	for acct, credsList := range creds2list {
		c.UI.Info(prettify(acct, credsList))
	}
	return 0
}

func organize(cred *models.GenericWrap) map[string]*[]*models.GenericCreds {
	organizedCreds := make(map[string]*[]*models.GenericCreds)
	for _, cred := range cred.Creds {
		if cred.SubType != models.SubCredType_HELM_REPO {
			continue
		}
		acctCreds := organizedCreds[cred.AcctName]
		if acctCreds == nil {
			acctCreds = &[]*models.GenericCreds{cred}
			organizedCreds[cred.AcctName] = acctCreds
			continue
		}
		*acctCreds = append(*acctCreds, cred)
	}
	return organizedCreds
}

func prettify(acctName string, envs *[]*models.GenericCreds) string {
	str := `Account: %s
Helm Repos:
%s
---
`
	var vars bytes.Buffer
	vars.WriteString("  ")
	for _, cred := range *envs {
		vars.WriteString(cred.Identifier + ": " + cred.ClientSecret + "\n  ")
	}
	return fmt.Sprintf(str, acctName, vars.String())
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return help
}

func Header(ui cli.Ui) {
	ui.Output("--- Helm Repo Credentials ---\n")
}

func NoDataHeader(ui cli.Ui) {
	ui.Warn("--- No Helm Repo Credentials Found! ---")
}

const synopsis = "List all environment variables held by ocelot"
const help = `
Usage: ocelot creds helmrepo list <options>

  Retrieves all helm repositories associated by each account. If you wish to filter the account variables to see, run: 
    ocelot creds helmrepo list -account <ACCT_NAME>
`

package repocredslist

import (
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/client/commandhelper"
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mitchellh/cli"
	"text/template"
)

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui, config: commandhelper.Config}
	c.init()
	return c
}

type cmd struct {
	UI cli.Ui
	flags   *flag.FlagSet
	accountFilter string
	config *commandhelper.ClientConfig
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
	msg, err := c.config.Client.GetRepoCreds(ctx, &protoReq)
	if err != nil {
		c.UI.Error(fmt.Sprint("Could not get list of credentials!\n Error: ", err.Error()))
	}
	printed := false
	Header(c.UI)
	for _, oneline := range msg.Repo {
		if c.accountFilter == "" || oneline.AcctName == c.accountFilter {
			c.UI.Info(Prettify(oneline))
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
	ui.Info("--- Repo Credentials ---\n")
}

func NoDataHeader(ui cli.Ui) {
	ui.Warn("--- No Repo Credentials Found! ---")
}

func Prettify(cred *models.RepoCreds) string {
	pretty := `Username: {{.Username}}
Password: {{.Password}}'
RepoUrls: {{ range $name, $url := .RepoUrl }}
   {{ $name }}: {{ $url }} {{ end }}
AcctName: {{.AcctName}}
Type: {{.Type}}
`
	fallback := `Username: %s
Password: %s
RepoUrl: %s
AcctName: %s
Type: %s
`
	tmpl, err := template.New("pretty").Parse(pretty)
	if err != nil {
		return fmt.Sprintf(fallback, cred.Username, cred.Password, cred.RepoUrl, cred.AcctName, cred.Type)
	}
	var here []byte
	buff := bytes.NewBuffer(here)
	err = tmpl.Execute(buff, cred)
	if err != nil {
		return fmt.Sprintf(fallback, cred.Username, cred.Password, cred.RepoUrl, cred.AcctName, cred.Type)
	}
	return buff.String()
}


const synopsis = "List all credentials used for artifact repositories"
const help = `
Usage: ocelot creds repo list

  Retrieves all credentials that ocelot uses to auth into artifact repositories
`

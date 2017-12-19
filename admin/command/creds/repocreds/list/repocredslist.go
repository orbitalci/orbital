package repocredslist

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
	msg, err := c.client.GetRepoCreds(ctx, &protoReq)
	if err != nil {
		c.UI.Error(fmt.Sprint("Could not get list of credentials!\n Error: ", err.Error()))
	}
	Header(c.UI)
	for _, oneline := range msg.Credentials {
		c.UI.Info(Prettify(oneline))
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
	pretty := `Username: %s
Password: %s
RepoUrl: %s
AcctName: %s
Type: %s

`
	return fmt.Sprintf(pretty, cred.Username, cred.Password, cred.RepoUrl, cred.AcctName, cred.Type)
}


const synopsis = "List all credentials used for artifact repositories"
const help = `
Usage: ocelot creds list

  Retrieves all credentials that ocelot uses to auth into artifact repositories
`

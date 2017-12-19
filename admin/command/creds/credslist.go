package creds

import (
	"bitbucket.org/level11consulting/ocelot/admin"
	"bitbucket.org/level11consulting/ocelot/admin/command/creds/buildcreds/list"
	"bitbucket.org/level11consulting/ocelot/admin/command/creds/repocreds/list"
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
	msg, err := c.client.GetAllCreds(ctx, &protoReq)
	if err != nil {
		c.UI.Error(fmt.Sprint("Could not get list of credentials!\n Error: ", err.Error()))
	}
	if len(msg.RepoCreds.Credentials) > 0 {
		repocredslist.Header(c.UI)
		for _, oneline := range msg.RepoCreds.Credentials {
			c.UI.Info(repocredslist.Prettify(oneline))
		}
	} else {
		repocredslist.NoDataHeader(c.UI)
	}

	if len(msg.AdminCreds.Credentials) > 0 {
		buildcredslist.Header(c.UI)
		for _, oneline :=  range msg.AdminCreds.Credentials {
			c.UI.Info(buildcredslist.Prettify(oneline))
		}
	} else {
		buildcredslist.NoDataHeader(c.UI)
	}

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

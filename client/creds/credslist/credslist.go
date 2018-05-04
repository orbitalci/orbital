package credslist

import (
	"context"
	"flag"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mitchellh/cli"
	"github.com/shankj3/ocelot/client/commandhelper"
	"github.com/shankj3/ocelot/client/creds/repocreds/list"
	"github.com/shankj3/ocelot/client/creds/vcscreds/list"
	models "github.com/shankj3/ocelot/models/pb"
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
}

func (c *cmd) Run(args []string) int {
	ctx := context.Background()
	var protoReq empty.Empty
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}
	msg, err := c.config.Client.GetAllCreds(ctx, &protoReq)
	if err != nil {
		c.UI.Error(fmt.Sprint("Could not get list of credentials!\n Error: ", err.Error()))
		return 1
	}
	if len(msg.RepoCreds.Repo) > 0 {
		repocredslist.Header(c.UI)
		for _, oneline := range msg.RepoCreds.Repo {
			c.UI.Output(repocredslist.Prettify(oneline))
		}
	} else {
		repocredslist.NoDataHeader(c.UI)
	}

	if len(msg.VcsCreds.Vcs) > 0 {
		buildcredslist.Header(c.UI)
		for _, oneline := range msg.VcsCreds.Vcs {
			c.UI.Output(buildcredslist.Prettify(oneline))
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

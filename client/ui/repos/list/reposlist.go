package reposlist

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mitchellh/cli"
	"github.com/olekukonko/tablewriter"
	"github.com/level11consulting/ocelot/client/commandhelper"
	models "github.com/level11consulting/ocelot/models/pb"
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
	msg, err := c.config.Client.GetTrackedRepos(ctx, &protoReq)
	if err != nil {
		c.UI.Error(fmt.Sprint("Could not get list of tracked repos!\n Error: ", err.Error()))
		return 1
	}

	switch len(msg.AcctRepos) {
	case 0:
		c.UI.Warn("--- No tracked repositories found ---\n")
	default:
		writer := &bytes.Buffer{}
		writ := tablewriter.NewWriter(writer)
		writ.SetAlignment(tablewriter.ALIGN_LEFT)
		writ.SetHeader([]string{"Account", "Repo", "Last Queued"})
		for _, acctrepo := range msg.AcctRepos {
			var unixTimeFormatted = "N/A"
			if acctrepo.LastQueue != nil {
				unixTimeFormatted = time.Unix(acctrepo.LastQueue.Seconds, int64(acctrepo.LastQueue.Nanos)).Format("Mon Jan _2 15:04:05")
			}
			var row []string
			row = append(row,
				acctrepo.Account,
				acctrepo.Repo,
				unixTimeFormatted,
			)
			writ.Append(row)
		}
		writ.Render()
		c.UI.Output("\n" + writer.String())
	}

	return 0
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return help
}

const synopsis = "list all repositories added to ocelot"
const help = `
Usage: ocelot repos list

  Will list all repositories that are being tracked by ocelot.
`

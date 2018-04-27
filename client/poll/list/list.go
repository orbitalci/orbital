package polllist

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mitchellh/cli"
	"github.com/olekukonko/tablewriter"
	"github.com/shankj3/ocelot/client/commandhelper"
	models "github.com/shankj3/ocelot/models/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const synopsis = "list all repositories currently tracked by ocelot"
const help = `
Usage: ocelot poll list 
`

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui, config: commandhelper.Config, OcyHelper: &commandhelper.OcyHelper{}}
	c.init()
	return c
}

type cmd struct {
	UI     cli.Ui
	flags  *flag.FlagSet
	config *commandhelper.ClientConfig
	*commandhelper.OcyHelper
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

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return help
}

func (c *cmd) init() {

}

func (c *cmd) Run(args []string) int {
	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}
	pollz, err := c.config.Client.ListPolledRepos(ctx, &empty.Empty{})
	if err != nil {
		errg, ok := status.FromError(err)
		if !ok {
			c.UI.Error("Error from server, I'm sorry. " + err.Error())
			return 1
		}
		if errg.Code() == codes.NotFound {
			c.UI.Info("No repos are currently being tracked via polling at this time.")
			return 0
		} else {
			c.UI.Error("Unable to retrieve list of repos, I'm sorry. Error: " + errg.Message())
			return 1
		}
	}
	writer := &bytes.Buffer{}
	writ := tablewriter.NewWriter(writer)
	writ.SetAlignment(tablewriter.ALIGN_LEFT)
	writ.SetHeader([]string{"Acct/Repo", "Cron String", "Branches", "Last Polled"})
	for _, poll := range pollz.Polls {
		var row []string
		thyme := time.Unix(poll.LastCronTime.Seconds, int64(poll.LastCronTime.Nanos))
		row = append(row,
			fmt.Sprintf("%s/%s", poll.Account, poll.Repo),
			poll.Cron,
			poll.Branches,
			thyme.Format("01/02/06 15:04:05"),
		)
		writ.Append(row)
	}
	writ.Render()
	c.UI.Output("\n" + writer.String())
	return 0
}

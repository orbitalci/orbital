package summary

import (
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/client/commandhelper"
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"github.com/olekukonko/tablewriter"
	"strings"
	"time"
)

const synopsis = "show summary table of specific repo"
const help = `
Usage: ocelot summary -acct-repo <acct>/<repo>
  Retrieve summary table of a specific repo (i.e. level11consulting/ocelot). If -limit is not specified, then the 
  limit will be set to 5, and only the last 5 runs will be shown.
  Full usage:
    $ ocelot summary -acct-repo jessishank/mytestocy -limit 2
    +----------+------------+-----------+--------+--------+--------------------+---------------------+----------+
    |  HASH    |  ACCOUNT   |   REPO    | BRANCH | FAILED | BUILD DURATION (S) |     START TIME      | BUILD ID |
    +----------+------------+-----------+--------+--------+--------------------+---------------------+----------+
    | ..75e9.. | jessishank | mytestocy | master | false  |             14.120 | Wed Jan 17 08:32:27 |       70 |
    | ..7860.. | jessishank | mytestocy | master | false  |             14.249 | Wed Jan 17 08:29:12 |       69 |
    +----------+------------+-----------+--------+--------+--------------------+---------------------+----------+

`


func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui, config: commandhelper.Config}
	c.init()
	return c
}

type cmd struct {
	UI cli.Ui
	flags   *flag.FlagSet
	config *commandhelper.ClientConfig
	accountRepo string
	limit int
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
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.StringVar(&c.accountRepo, "acct-repo", "ERROR", "*REQUIRED*  <account>/<repo> to track")
	c.flags.IntVar(&c.limit, "limit", 5, "number of rows to fetch")
}


func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	if c.accountRepo == "ERROR" {
		c.UI.Error("flag -acct-repo must be in the format <account>/<repo>. see --help")
		return 1
	}
	data := strings.Split(c.accountRepo, "/")
	if len(data) != 2  {
		c.UI.Error("flag -acct-repo must be in the format <account>/<repo>. see --help")
		return 1
	}
	account := data[0]
	repo := data[1]
	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}
	summaries, err := c.config.Client.LastFewSummaries(ctx, &models.RepoAccount{Repo: repo, Account: account, Limit: int32(c.limit)})
	if err != nil {
		// todo: add more descriptive error
		c.UI.Error("unable to get build summaries! error: " + err.Error())
		return 1
	}
	writer := &bytes.Buffer{}
	writ := tablewriter.NewWriter(writer)
	writ.SetHeader([]string{"Hash", "Account", "Repo", "Branch", "Failed", "Build Duration (s)", "Start Time", "Build ID"})

	for _, sum := range summaries.Sums {
		writ.Append(generateTableRow(sum))
	}
	writ.Render()
	c.UI.Info(writer.String())
	return 0
}

func generateTableRow(summary *models.BuildSummary) []string {
	tym := time.Unix(summary.BuildTime.Seconds, int64(summary.BuildTime.Nanos))
	row := []string{
		summary.Hash,
		summary.Account,
		summary.Repo,
		summary.Branch,
		fmt.Sprintf("%t", summary.Failed),
		fmt.Sprintf("%.3f", summary.BuildDuration),
		tym.Format("Mon Jan 2 15:04:05"),
		fmt.Sprintf("%d", summary.BuildId),
	}
	return row
}
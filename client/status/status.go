package status

import (
	"bitbucket.org/level11consulting/ocelot/client/commandhelper"
	"github.com/mitchellh/cli"
	"flag"
	"strings"
	"context"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"fmt"
	"bitbucket.org/level11consulting/ocelot/util/cmd_table"
	"github.com/golang/protobuf/ptypes/wrappers"
	"bytes"
	"github.com/olekukonko/tablewriter"
	"time"
)

const synopsis = "show status of specific acctname, acctname/repo, or hash"
//TODO: finish writing help once I know how it works
const help = `
Usage: ocelot status -acct-repo <acct>/<repo>
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
	account string
	hash string
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
	//we accept all 3 flags, but prioritize output in the following order: hash, acct-repo, acct
	c.flags.StringVar(&c.hash, "hash", "ERROR", "[optional]  <hash> to display build status")
	c.flags.StringVar(&c.account, "acct", "ERROR", "[optional]  <account> to display build status")
	c.flags.StringVar(&c.accountRepo, "acct-repo", "ERROR", "[optional]  <account>/<repo> to display build status")
}

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}

	if c.accountRepo == "ERROR" && c.account == "ERROR" && c.hash == "ERROR" {
		c.UI.Error("one of the following flags must be set: -acct-repo, -acct, -hash. see --help")
		return 1
	}


	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}

	// always respect hash first
	if c.hash != "ERROR" {
		builds, err := c.GetClient().BuildRuntime(ctx, &models.BuildQuery{
			Hash: c.hash,
		})
		if err != nil {
			c.UI.Error(fmt.Sprintf("error retrieving build runtime for hash %s. Error: %s", c.hash, err.Error()))
			return 1
		}
		if len(builds.Builds) == 0 {
			c.UI.Info(fmt.Sprintf("no data found for hash %s", c.hash))
			return 0
		} else if len(builds.Builds) > 1 {
			c.UI.Info(fmt.Sprintf("%d hashes matching `%s`, please select a more complete git hash: ", len(builds.Builds), c.hash))
			c.UI.Output(cmd_table.SelectFromHashes(builds))
			return 0
		}


		//it's okay to iterate here cause list will always contain 1 value
		for _, build := range builds.Builds {
			//TODO: review w/ somebody - should I just change RetrieveLatestSum to take in partial hash and then we can avoid additional client call?
			hashStages, err := c.GetClient().StatusByHash(ctx, &wrappers.StringValue{Value: build.Hash})
			if err != nil {
				c.UI.Error(fmt.Sprintf("error retrieving status for hash %s. Error: %s", c.hash, err.Error()))
				return 1
			}

			c.UI.Output(fmt.Sprintf("build id %s", hashStages.BuildSum.BuildId))
			writer := &bytes.Buffer{}
			writ := tablewriter.NewWriter(writer)
			writ.SetAlignment(tablewriter.ALIGN_LEFT)   // Set Alignment
			writ.SetHeader([]string{"Build ID", "Start Time", "Stage Duration", "Status", "Stage", "Messages"})
			writ.SetHeaderColor(
				tablewriter.Colors{tablewriter.FgBlackColor, tablewriter.Bold},
				tablewriter.Colors{tablewriter.FgBlackColor, tablewriter.Bold},
				tablewriter.Colors{tablewriter.FgBlackColor, tablewriter.Bold},
				tablewriter.Colors{tablewriter.FgBlackColor, tablewriter.Bold},
				tablewriter.Colors{tablewriter.FgBlackColor, tablewriter.Bold},
				tablewriter.Colors{tablewriter.FgBlackColor, tablewriter.Bold})


			for _, stage := range hashStages.Stages {
				tym := time.Unix(stage.StartTime.Seconds, int64(stage.StartTime.Nanos))
				writ.Append([]string{
					fmt.Sprintf("%v", hashStages.BuildSum.BuildId),
					tym.Format("Mon Jan 2 15:04:05"),
					commandhelper.PrettifyTime(stage.StageDuration),
					fmt.Sprintf("%v", stage.Status),
					stage.Stage,
					strings.Join(stage.Messages, "\n")})
			}

			writ.Render()
			c.UI.Output("\n" + writer.String())
			return 0
		}


		return 0
	}

	//respect acct-repo next

	//acct is last


	data := strings.Split(c.accountRepo, "/")
	if len(data) != 2  {
		c.UI.Error("flag -acct-repo must be in the format <account>/<repo>. see --help")
		return 1
	}



	return 0
}
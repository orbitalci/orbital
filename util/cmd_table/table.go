package cmd_table

import (
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bytes"
	"github.com/olekukonko/tablewriter"
	"fmt"
	"bitbucket.org/level11consulting/ocelot/client/commandhelper"
	"strings"
)

// cmd_table package contains utils for drawing tables



//SelectFromHashes will draw a table that can be displayed if there's multiple matching hashes
//+------------------------------------------+----------------------------+-------------------+
//|                   HASH                   |            REPO            |   ACCOUNT NAME    |
//+------------------------------------------+----------------------------+-------------------+
//| ee                                       | ---                        | ---               |
//| ec8ea5f46cdd198c135c1ba73984ac6d6192cc16 | orchestr8-locationservices | level11consulting |
//+------------------------------------------+----------------------------+-------------------+
//It takes in ocelot server's BuildRuntime response
func SelectFromHashes(build *models.Builds) string {
	writer := &bytes.Buffer{}
	writ := tablewriter.NewWriter(writer)
	writ.SetAlignment(tablewriter.ALIGN_LEFT)   // Set Alignment
	writ.SetHeader([]string{"Hash", "Repo", "Account Name"})
	writ.SetHeaderColor(
		tablewriter.Colors{tablewriter.FgBlackColor, tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgBlackColor, tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgBlackColor, tablewriter.Bold})

	for _, build := range build.Builds {
		var buildLine []string
		buildLine = append(buildLine, build.Hash)
		repoName := build.RepoName
		acctName := build.AcctName

		if len(repoName) == 0 {
			repoName = "---"
		}

		if len(acctName) == 0 {
			acctName = "---"
		}

		buildLine = append(buildLine, repoName)
		buildLine = append(buildLine, acctName)

		writ.Append(buildLine)
	}

	writ.Render()
	explanationMsg := fmt.Sprintf("\033[0;34mit's your lucky day, there's %d hashes matching that value. Please enter a more complete git hash\n\033[m", len(build.Builds))
	return explanationMsg + writer.String()
}

//PrintStatusTable is used in printing out status-es such like this:
//
//mariannefeng/test-ocelot 	 496b3e3a9b8b6f773a4562a50bb30863471e2adf 	 PASS
//	setup in 3 seconds
//	build in 17 seconds
//
//it takes in a boolean argument indicating whether or not the build is running, and a protobuf Status
//object. It returns a PASS/FAIL/Running status string, a color corresponding with that status,
//and the string representation of stages, stage messages, and errors if exists
func PrintStatusStages(isRunning bool, statuses *models.Status) (string, int, string) {
	var status, stageStatus string
	var color int

	if isRunning {
		status = "Running"
		color = 33
	} else if statuses != nil {
		if !statuses.BuildSum.Failed {
			status = "PASS"
			color = 32
		} else {
			status = "FAIL"
			color = 31
		}
	}

	if statuses != nil {
		for _, stage := range statuses.Stages {
			var stageStatusStr string
			if stage.Status == 0 {
				stageStatusStr = "PASS"
			} else {
				stageStatusStr = "FAIL"
			}
			stageStatus += fmt.Sprintf("\n[%s] took %s to %s", stage.Stage, commandhelper.PrettifyTime(stage.StageDuration), stageStatusStr)
			if statuses.BuildSum.Failed {
				stageStatus += fmt.Sprintf("\n\t * %s", strings.Join(stage.Messages, "\n\t * "))
				if len(stage.Error) > 0 {
					stageStatus += fmt.Sprintf(": \033[1;30m%s\033[0m", stage.Error)
				}
			}
		}
	}
	return stageStatus + "\n", color, status
}

func PrintStatusOverview(color int, acctName, repoName, hash, status string) string {
	buildStatus := fmt.Sprintf("\n\033[1;%dmstatus: %s\033[0m \n\033[0;33mhash: %s\033[0m\naccount: %s \nrepo: %s\n", color, status, hash, acctName, repoName)
	return buildStatus
}
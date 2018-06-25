package commandhelper

import (
	"bytes"
	"fmt"
	"github.com/olekukonko/tablewriter"
	models "github.com/shankj3/ocelot/models/pb"
	"os"
	"strings"
)

// package contains utils for drawing tables


//SelectFromHashes will draw a table that can be displayed if there's multiple matching hashes
//+------------------------------------------+----------------------------+-------------------+
//|                   HASH                   |            REPO            |   ACCOUNT NAME    |
//+------------------------------------------+----------------------------+-------------------+
//| ee                                       | ---                        | ---               |
//| ec8ea5f46cdd198c135c1ba73984ac6d6192cc16 | orchestr8-locationservices | level11consulting |
//+------------------------------------------+----------------------------+-------------------+
//It takes in ocelot server's BuildRuntime response
func SelectFromHashes(build *models.Builds, theme *ColorDefs) string {
	writer := &bytes.Buffer{}
	writ := tablewriter.NewWriter(writer)
	writ.SetAlignment(tablewriter.ALIGN_LEFT) // Set Alignment
	writ.SetHeader([]string{"Hash", "Repo", "Account Name"})
	if !theme.NoColor {
		writ.SetHeaderColor(
			tablewriter.Colors{tablewriter.FgBlackColor, tablewriter.Bold},
			tablewriter.Colors{tablewriter.FgBlackColor, tablewriter.Bold},
			tablewriter.Colors{tablewriter.FgBlackColor, tablewriter.Bold})
	}

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
	explanationMsg := theme.Info.Sprintf("it's your lucky day, there's %d hashes matching that value. Please enter a more complete git hash\n", len(build.Builds))
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
func PrintStatusStages(statuses *models.Status, wide bool, theme *ColorDefs) (string, *Color, string) {
	var status, stageStatus string
	var color *Color
	switch statuses.BuildSum.Status {
	case models.BuildStatus_RUNNING:
		status = "Running"
		color = theme.Running
	case models.BuildStatus_QUEUED:
		status = "Queued and waiting to be built"
		color = theme.Queued
		return stageStatus, color, status
	case models.BuildStatus_FAILED_PRESTART:
		status = "Failed PreStart"
		color = theme.Failed
	case models.BuildStatus_FAILED:
		status = "FAIL"
		color = theme.Failed
	case models.BuildStatus_PASSED:
		status = "PASS"
		color = theme.Passed
	default:
		theme.Error.Println("Status is nil, this should not happen.")
		os.Exit(1)
	    color = theme.Normal
	}

	if statuses != nil && len(statuses.Stages) > 0 {
		for _, stage := range statuses.Stages {
			var stageStatusStr string
			if stage.Status == 0 {
				stageStatusStr = "PASS"
			} else {
				stageStatusStr = "FAIL"
			}
			stageStatus += fmt.Sprintf("\n[%s] took %s to %s", stage.StageStatus, PrettifyTime(stage.StageDuration, statuses.BuildSum.Status==models.BuildStatus_QUEUED), stageStatusStr)
			if statuses.BuildSum.Failed || wide {
				stageStatus += fmt.Sprintf("\n\t * %s", strings.Join(stage.Messages, "\n\t * "))
				if len(stage.Error) > 0 {
					stageStatus += theme.Normal.Sprintf(": %s", stage.Error)
				}
			}
		}
		stageStatus += "\n"
	}
	return stageStatus, color, status
}

func PrintStatusOverview(color *Color, acctName, repoName, hash, status string, theme *ColorDefs) string {
	fmt.Println(color)
	buildStatus := color.Sprintf("\nstatus: %s ", status) + theme.Warning.Sprintf("\nhash: %s", hash) + fmt.Sprintf("\naccount: %s \nrepo: %s\n", acctName, repoName)
	return buildStatus
}

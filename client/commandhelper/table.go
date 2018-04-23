package commandhelper

import (
	models "bitbucket.org/level11consulting/ocelot/models/pb"
	"bytes"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"strings"
)

// package contains utils for drawing tables

type BuildStatus int

const (
	RUNNING BuildStatus = iota
	QUEUED
	DONE
	FAILED_PRESTART
)

//queued := statuses.BuildSum.BuildTime.Seconds == 0 && statuses.BuildSum.BuildTime.Nanos == 0
//buildStarted := statuses.BuildSum.BuildTime.Seconds > 0 && statuses.IsInConsul
//finished := !statuses.IsInConsul && buildStarted
func GetStatus(queued, buildStarted, finished, failed_validation bool) BuildStatus {
	if queued { return QUEUED }
	if buildStarted { return RUNNING }
	if finished { return DONE }
	if failed_validation { return FAILED_PRESTART }
	panic("none of these!")
}

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
	explanationMsg := fmt.Sprintf("%sit's your lucky day, there's %d hashes matching that value. Please enter a more complete git hash\n%s", theme.Info, len(build.Builds), theme.Reset)
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
func PrintStatusStages(bs BuildStatus, statuses *models.Status, wide bool, theme *ColorDefs) (string, string, string) {
	var status, stageStatus string
	var color string
	switch bs {
	case RUNNING:
		status = "Running"
		color = theme.Running
	case QUEUED:
		status = "Queued and waiting to be built"
		color = theme.Queued
		return stageStatus, color, status
	case FAILED_PRESTART:
		status = "Failed PreStart"
		color = theme.Failed
	case DONE:
		if !statuses.BuildSum.Failed {
			status = "PASS"
			color = theme.Passed
		} else {
			status = "FAIL"
			color = theme.Failed
		}
	}

	if statuses != nil && len(statuses.Stages) > 0 {
		for _, stage := range statuses.Stages {
			var stageStatusStr string
			if stage.Status == 0 {
				stageStatusStr = "PASS"
			} else {
				stageStatusStr = "FAIL"
			}
			stageStatus += fmt.Sprintf("\n[%s] took %s to %s", stage.StageStatus, PrettifyTime(stage.StageDuration, bs == QUEUED), stageStatusStr)
			if statuses.BuildSum.Failed || wide {
				stageStatus += fmt.Sprintf("\n\t * %s", strings.Join(stage.Messages, "\n\t * "))
				if len(stage.Error) > 0 {
					stageStatus += fmt.Sprintf(": %s%s%s", theme.Normal, stage.Error, theme.Reset)
				}
			}
		}
		stageStatus += "\n"
	}
	return stageStatus, color, status
}

func PrintStatusOverview(color int, acctName, repoName, hash, status string, theme *ColorDefs) string {
	buildStatus := fmt.Sprintf("\n%sstatus: %s%s \n%shash: %s%s\naccount: %s \nrepo: %s\n", color, status,  theme.Reset, theme.Warning, hash, theme.Reset, acctName, repoName)
	return buildStatus
}


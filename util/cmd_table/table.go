package cmd_table

import (
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bytes"
	"github.com/olekukonko/tablewriter"
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
	return writer.String()
}
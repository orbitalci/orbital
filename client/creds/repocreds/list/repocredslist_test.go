package repocredslist


import (
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/client/commandhelper"
	"context"
	"flag"
	"github.com/mitchellh/cli"
	"strings"
	"testing"
)

func TestCmd_Run(t *testing.T) {
	ctx := context.Background()
	ui := cli.NewMockUi()
	cmdd := &cmd{
		UI: ui,
		config: commandhelper.NewTestClientConfig([]string{}),
	}
	cmdd.flags = flag.NewFlagSet("", flag.ContinueOnError)
	cmdd.flags.StringVar(&cmdd.accountFilter, "account", "",
		"")
	expectedCreds := &models.RepoCredWrapper{
		Repo: []*models.RepoCreds{
			{
				Username:     "thisBeMyUserName",
				Password:     "SHH-BE-QUIET-ITS-A-SECRET",
				RepoUrl:      "https://ocelot.perf/nexus-yo",
				AcctName:     "jessishank",
				Type:         "nexus",
			},
			{
				Username:     "thisBeM1yUserName",
				Password:     "SHH-BE-Q2UIET-ITS-A-SECRET",
				RepoUrl:      "https://o3celot.perf/nexus-yo",
				AcctName:     "jessishank45",
				Type:         "nexus",
			},{
				Username:     "thisB2eMyUserName",
				Password:     "SHH-BEd-QUIET-asdITS-A-SECRET",
				RepoUrl:      "https:/h/ocelot.perf/nexus-yo",
				AcctName:     "jessishasnk",
				Type:         "nexus",
			},
		},
	}

	for _, cred := range expectedCreds.Repo {
		cmdd.config.Client.SetRepoCreds(ctx, cred)
	}
	var args []string
	if exit := cmdd.Run(args); exit != 0 {
		t.Error("should exit with code 0, exited with code ", exit)
	}

	//_, err := ui.ErrorWriter.Read(stdout)
	//if err != nil {
	//	t.Fatal("could not read stdout from buffer")
	//}
	expectedText := `--- Repo Credentials ---

Username: thisBeMyUserName
Password: SHH-BE-QUIET-ITS-A-SECRET
RepoUrl: https://ocelot.perf/nexus-yo
AcctName: jessishank
Type: nexus


Username: thisBeM1yUserName
Password: SHH-BE-Q2UIET-ITS-A-SECRET
RepoUrl: https://o3celot.perf/nexus-yo
AcctName: jessishank45
Type: nexus


Username: thisB2eMyUserName
Password: SHH-BEd-QUIET-asdITS-A-SECRET
RepoUrl: https:/h/ocelot.perf/nexus-yo
AcctName: jessishasnk
Type: nexus


`
	text := ui.OutputWriter.String()
	if strings.Compare(expectedText, text) != 0 {
		t.Errorf("output and expected not the same,  \n" +
			"expected:\n%s\ngot:\n%s", expectedText, text)
	}


}

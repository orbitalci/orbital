package buildcredslist

import (
	"context"
	"flag"
	"strings"
	"testing"

	"bitbucket.org/level11consulting/ocelot/client/commandhelper"
	models "bitbucket.org/level11consulting/ocelot/models/pb"
	"github.com/mitchellh/cli"
)

func testNew() (*cmd, *cli.MockUi) {
	ui := cli.NewMockUi()
	c := &cmd{
		UI: ui,
		config: commandhelper.NewTestClientConfig([]string{}),
	}
	return c, ui
}

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
	expectedCreds := &models.CredWrapper{
		Vcs: []*models.VCSCreds{
			{
				ClientId:     "fancy-frickin-identification",
				ClientSecret: "SHH-BE-QUIET-ITS-A-SECRET",
				TokenURL:     "https://ocelot.perf/site/oauth2/access_token",
				AcctName:     "lamb-shank",
				Identifier:   "howdy",
				SubType:      models.SubCredType_BITBUCKET,
			},
			{
				ClientId:     "fancy-trickin-identification",
				ClientSecret: "SHH-BEE-QUIET-ITS-A-SECRET",
				TokenURL:     "https://oqelot.perf/site/oauth2/access_token",
				AcctName:     "lamf-shank",
				Identifier:   "rowdy",
				SubType:      models.SubCredType_GITHUB,
			},
		},
	}

	for _, cred := range expectedCreds.Vcs {
		cmdd.config.Client.SetVCSCreds(ctx, cred)
	}
	var args []string
	if exit := cmdd.Run(args); exit != 0 {
		t.Error("should exit with code 0, exited with code ", exit)
	}

	//_, err := ui.ErrorWriter.Read(stdout)
	//if err != nil {
	//	t.Fatal("could not read stdout from buffer")
	//}
	expectedText := `
--- Admin Credentials ---

ClientId: fancy-frickin-identification
ClientSecret: SHH-BE-QUIET-ITS-A-SECRET
TokenURL: https://ocelot.perf/site/oauth2/access_token
AcctName: lamb-shank
SubType: bitbucket
Identifier: howdy
[THIS IS A TEST]


ClientId: fancy-trickin-identification
ClientSecret: SHH-BEE-QUIET-ITS-A-SECRET
TokenURL: https://oqelot.perf/site/oauth2/access_token
AcctName: lamf-shank
SubType: github
Identifier: rowdy
[THIS IS A TEST]


`
	text := ui.OutputWriter.String()
	if strings.Compare(expectedText, text) != 0 {
		t.Errorf("output and expected not the same,  \n" +
			"expected:\n%s-----\ngot:\n%s-----", expectedText, text)
	}


}
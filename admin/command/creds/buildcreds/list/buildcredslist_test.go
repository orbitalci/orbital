package buildcredslist

import (
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"context"
	"github.com/mitchellh/cli"
	"flag"
	"strings"
	"testing"
)

func testNew() (*cmd, *cli.MockUi) {
	ui := cli.NewMockUi()
	c := &cmd{
		UI: ui,
		client: models.NewFakeGuideOcelotClient(),
	}
	return c, ui
}

func TestCmd_Run(t *testing.T) {
	ctx := context.Background()
	ui := cli.NewMockUi()
	cmdd := &cmd{
		UI: ui,
		client: models.NewFakeGuideOcelotClient(),
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
				Type:         "bitbucket",
			},
			{
				ClientId:     "fancy-trickin-identification",
				ClientSecret: "SHH-BEE-QUIET-ITS-A-SECRET",
				TokenURL:     "https://oqelot.perf/site/oauth2/access_token",
				AcctName:     "lamf-shank",
				Type:         "github",
			},
		},
	}

	for _, cred := range expectedCreds.Vcs {
		cmdd.client.SetVCSCreds(ctx, cred)
	}
	var args []string
	if exit := cmdd.Run(args); exit != 0 {
		t.Error("should exit with code 0, exited with code ", exit)
	}

	//_, err := ui.ErrorWriter.Read(stdout)
	//if err != nil {
	//	t.Fatal("could not read stdout from buffer")
	//}
	expectedText := `--- Admin Credentials ---

ClientId: fancy-frickin-identification
ClientSecret: SHH-BE-QUIET-ITS-A-SECRET
TokenURL: https://ocelot.perf/site/oauth2/access_token
AcctName: lamb-shank
Type: bitbucket


ClientId: fancy-trickin-identification
ClientSecret: SHH-BEE-QUIET-ITS-A-SECRET
TokenURL: https://oqelot.perf/site/oauth2/access_token
AcctName: lamf-shank
Type: github


`
	text := ui.OutputWriter.String()
	if strings.Compare(expectedText, text) != 0 {
		t.Errorf("output and expected not the same,  \n" +
			"expected:\n%s\ngot:\n%s", expectedText, text)
	}


}
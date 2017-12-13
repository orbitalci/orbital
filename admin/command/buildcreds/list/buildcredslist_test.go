package buildcredslist

import (
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"context"
	"github.com/mitchellh/cli"
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
	cmdd, ui := testNew()
	expectedCreds := &models.CredWrapper{
		Credentials: []*models.Credentials{
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
	var expectedText string
	for _, cred := range expectedCreds.Credentials {
		expectedText += prettify(cred)
		cmdd.client.SetCreds(ctx, cred)
	}
	if exit := cmdd.Run([]string{""}); exit != 0 {
		t.Error("should exit with code 0, exited with code ", exit)
	}
	var stdout []byte
	//_, err := ui.ErrorWriter.Read(stdout)
	//if err != nil {
	//	t.Fatal("could not read stdout from buffer")
	//}
	if expectedText != ui.ErrorWriter.String() {
		t.Errorf("output and expected not the same,  \nexpected:\n %s \n\n got:\n%s", expectedText, string(stdout))
	}


}
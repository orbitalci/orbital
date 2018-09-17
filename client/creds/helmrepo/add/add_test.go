package helmrepoadd

import (
	"bytes"
	"testing"

	"github.com/go-test/deep"
	"github.com/mitchellh/cli"
	"github.com/shankj3/go-til/test"
	"github.com/shankj3/ocelot/common/testutil"
	"github.com/shankj3/ocelot/models/pb"

	"github.com/shankj3/ocelot/client/commandhelper"
)

func TestCmd_Run_exists(t *testing.T) {
	ui := cli.NewMockUi()
	ui.InputReader = bytes.NewReader([]byte("YES"))
	conf := commandhelper.NewTestClientConfig([]string{})
	c := &cmd{
		UI:     ui,
		config: conf,
	}
	c.init()
	client := c.config.Client.(*testutil.FakeGuideOcelotClient)
	client.Generics = &pb.GenericWrap{}
	client.Exists = true
	if code := c.Run([]string{"-acct=jessishank", "-repo-name=ocelot_charts", "-helm-url=https://ocelot.github.io/helm"}); code != 0 {
		t.Error("shouldn't fail")
	}
	expected := `The helm repo address with the name ocelot_charts already exists under the account jessishank. Do you wish to overwrite? Only a YES will continue and update the helm chart repo value, otherwise this entry will be skipped.`
	out := ui.OutputWriter.String()
	if expected != out {
		t.Error(test.StrFormatErrors("question msg", expected, out))
	}
	if len(client.Generics.Creds) != 1 {
		t.Error("should have uploaded credentials")
	}
	ui = cli.NewMockUi()
	client.Generics.Creds = []*pb.GenericCreds{}
	ui.InputReader = bytes.NewReader([]byte("potato"))
	c.UI = ui
	if code := c.Run([]string{"-acct=jessishank", "-repo-name=ocelot_charts", "-helm-url=https://ocelot.github.io/helm"}); code != 0 {
		t.Error("shouldn't fail")
	}
	if len(client.Generics.Creds) != 0 {
		t.Error("didn't answer yes, should have not uploaded credential.")
	}

}

func TestCmd_Run_noparams(t *testing.T) {
	ui := cli.NewMockUi()
	conf := commandhelper.NewTestClientConfig([]string{})
	c := &cmd{
		UI:     ui,
		config: conf,
	}
	c.init()
	client := c.config.Client.(*testutil.FakeGuideOcelotClient)
	client.Generics = &pb.GenericWrap{}
	if code := c.Run([]string{"-acct=jessishank"}); code != 1 {
		t.Error("should fail")
	}
}

func TestCmd_Run(t *testing.T) {
	ui := cli.NewMockUi()
	conf := commandhelper.NewTestClientConfig([]string{})
	c := &cmd{
		UI:     ui,
		config: conf,
	}
	c.init()
	client := c.config.Client.(*testutil.FakeGuideOcelotClient)
	client.Generics = &pb.GenericWrap{}
	if code := c.Run([]string{"-acct=jessishank", "-repo-name=ocelot_charts", "-helm-url=https://ocelot.github.io/helm"}); code != 0 {
		t.Error("shouldn't fail")
	}
	expected := &pb.GenericWrap{
		Creds: []*pb.GenericCreds{
			{
				AcctName:     "jessishank",
				Identifier:   "ocelot_charts",
				ClientSecret: "https://ocelot.github.io/helm",
				SubType:      pb.SubCredType_HELM_REPO,
			},
		},
	}
	if diff := deep.Equal(expected, client.Generics); diff != nil {
		t.Error(diff)
	}
}

func TestCmd_Run_badArg(t *testing.T) {
	ui := cli.NewMockUi()
	conf := commandhelper.NewTestClientConfig([]string{})
	c := &cmd{
		UI:     ui,
		config: conf,
	}
	c.init()
	if code := c.Run([]string{"peanuckle"}); code != 1 {
		t.Error("should fail, bad arg")
	}
}

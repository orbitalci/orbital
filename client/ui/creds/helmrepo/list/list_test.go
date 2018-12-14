package helmrepolist

import (
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/mitchellh/cli"
	"github.com/shankj3/go-til/test"
	"github.com/level11consulting/orbitalci/client/commandhelper"
	"github.com/level11consulting/orbitalci/common/testutil"
	"github.com/level11consulting/orbitalci/models/pb"
)

var creds = []*pb.GenericCreds{
	{
		AcctName:     "one",
		Identifier:   "urlco",
		ClientSecret: "https://theurl.co",
		SubType:      pb.SubCredType_HELM_REPO,
	},
	{
		AcctName:     "one",
		Identifier:   "helmco",
		ClientSecret: "http://helm.co",
		SubType:      pb.SubCredType_HELM_REPO,
	},
	{
		AcctName:     "two",
		Identifier:   "identifier3",
		ClientSecret: "https://identifier.ci",
		SubType:      pb.SubCredType_HELM_REPO,
	},
	{
		AcctName:     "two",
		Identifier:   "identifier3",
		ClientSecret: "https://identifier.ci",
		SubType:      pb.SubCredType_ENV,
	},
}

func Test_organize(t *testing.T) {
	organized := organize(&pb.GenericWrap{Creds: creds})
	onecreds, ok := organized["one"]
	if !ok {
		t.Error("should have an array of helm_repo creds under acctname 'one'")
	}
	knownCreds := []*pb.GenericCreds{creds[0], creds[1]}
	if diff := deep.Equal(knownCreds, *onecreds); diff != nil {
		t.Error(diff)
	}
	twocreds, ok := organized["two"]
	if !ok {
		t.Error("should have an array of env creds under acctname 'two'")
	}
	known2Creds := []*pb.GenericCreds{creds[2]}
	if diff := deep.Equal(known2Creds, *twocreds); diff != nil {
		t.Error(diff)
	}
}

func TestCmd_Run_badArgs(t *testing.T) {
	ui := cli.NewMockUi()
	conf := commandhelper.NewTestClientConfig([]string{})
	c := &cmd{
		UI:     ui,
		config: conf,
	}
	c.init()
	code := c.Run([]string{"-banana"})
	if code != 1 {
		t.Error("should return one, got " + string(code))
	}
}

func TestCmd_Run_notConnected(t *testing.T) {
	ui := cli.NewMockUi()
	conf := commandhelper.NewTestClientConfig([]string{})
	c := &cmd{
		UI:     ui,
		config: conf,
	}
	c.init()
	conf.Client.(*testutil.FakeGuideOcelotClient).NotConnected = true
	if code := c.Run([]string{}); code != 1 {
		t.Error("should have failed")
	}
	output := ui.ErrorWriter.String()
	if !strings.Contains(output, "Could not connect to server") {
		t.Error("should return connection error ")
	}
}

func TestCmd_Run_error(t *testing.T) {
	ui := cli.NewMockUi()
	conf := commandhelper.NewTestClientConfig([]string{})
	c := &cmd{
		UI:     ui,
		config: conf,
	}
	c.init()
	conf.Client.(*testutil.FakeGuideOcelotClient).ReturnError = true
	if code := c.Run([]string{}); code != 1 {
		t.Error("should have failed")
	}
	output := ui.ErrorWriter.String()
	if !strings.Contains(output, "Could not get list of credentials!") {
		t.Error("should have returned could not get creds error")
	}
}

func TestCmd_Run_filter(t *testing.T) {
	ui := cli.NewMockUi()
	conf := commandhelper.NewTestClientConfig([]string{})
	c := &cmd{
		UI:     ui,
		config: conf,
	}
	c.init()
	conf.Client.(*testutil.FakeGuideOcelotClient).Exists = true
	conf.Client.(*testutil.FakeGuideOcelotClient).Generics = &pb.GenericWrap{Creds: creds}
	if code := c.Run([]string{"-account=monopoly"}); code != 0 {
		t.Error("should not have failed.")
	}
	output := ui.ErrorWriter.String()
	expected := `--- No Helm Repo Credentials Found! ---
`
	if output != expected {
		t.Error(test.StrFormatErrors("output", expected, output))
	}
	ui = cli.NewMockUi()
	c.UI = ui
	if code := c.Run([]string{"-account=one"}); code != 0 {
		t.Error("should not have failed.")
	}
	expected = `--- Helm Repo Credentials ---

Account: one
Helm Repos:
  urlco: https://theurl.co
  helmco: http://helm.co
  
---

`
	output = ui.OutputWriter.String()
	if expected != output {
		t.Error(test.StrFormatErrors("output", expected, output))
	}
}

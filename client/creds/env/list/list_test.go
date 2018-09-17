package envlist

import (
	"flag"
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/mitchellh/cli"
	"github.com/shankj3/go-til/test"
	"github.com/shankj3/ocelot/client/commandhelper"
	"github.com/shankj3/ocelot/common/testutil"
	"github.com/shankj3/ocelot/models/pb"
)

var creds = []*pb.GenericCreds{
	{
		AcctName:     "one",
		Identifier:   "identifier1",
		ClientSecret: "secret1",
		SubType:      pb.SubCredType_ENV,
	},
	{
		AcctName:     "one",
		Identifier:   "identifier2",
		ClientSecret: "secret2",
		SubType:      pb.SubCredType_ENV,
	},
	{
		AcctName:     "two",
		Identifier:   "identifier3",
		ClientSecret: "secret3",
		SubType:      pb.SubCredType_ENV,
	},
	{
		AcctName:     "one",
		Identifier:   "identifier4",
		ClientSecret: "secret4",
		SubType:      pb.SubCredType_ENV,
	},
}

func Test_organize(t *testing.T) {
	organized := organize(&pb.GenericWrap{Creds: creds})
	onecreds, ok := organized["one"]
	if !ok {
		t.Error("should have an array of env creds under acctname 'one'")
	}
	knownCreds := []*pb.GenericCreds{creds[0], creds[1], creds[3]}
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
	conf.Client.(*testutil.FakeGuideOcelotClient).Generics = &pb.GenericWrap{Creds: moreCreds}
	if code := c.Run([]string{"-account=monopoly"}); code != 0 {
		t.Error("should not have failed.")
	}
	output := ui.ErrorWriter.String()
	expected := `--- No Env Credentials Found! ---
`
	if output != expected {
		t.Error(test.StrFormatErrors("output", expected, output))
	}
	ui = cli.NewMockUi()
	c.UI = ui
	if code := c.Run([]string{"-account=shankj3"}); code != 0 {
		t.Error("should not have failed.")
	}
	expected = `--- Env Credentials ---

Account: shankj3
Env Vars:
  ENVIRONMENT=********
  EIINVIRONMENT=********
  whoopwhooOop=********
  yellowNotepad=********
  
---

`
	output = ui.OutputWriter.String()
	if expected != output {
		t.Error(test.StrFormatErrors("output", expected, output))
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
	conf.Client.(*testutil.FakeGuideOcelotClient).Exists = true
	conf.Client.(*testutil.FakeGuideOcelotClient).Generics = &pb.GenericWrap{Creds: []*pb.GenericCreds{{AcctName: "shankj3", Identifier: "ENVIRONMENT", ClientSecret: "********", SubType: pb.SubCredType_ENV}}}
	code := c.Run([]string{})
	if code != 0 {
		t.Error(code)
	}
	expected := `--- Env Credentials ---

Account: shankj3
Env Vars:
  ENVIRONMENT=********
  
---

`
	live := ui.OutputWriter.String()
	if expected != live {
		t.Error(test.StrFormatErrors("output", expected, live))
	}
	ui = cli.NewMockUi()
	c.UI = ui
	conf.Client.(*testutil.FakeGuideOcelotClient).Generics.Creds = moreCreds
	code = c.Run([]string{})
	if code != 0 {
		t.Error(code)
	}
	expected1 := `
Account: shankj3
Env Vars:
  ENVIRONMENT=********
  EIINVIRONMENT=********
  whoopwhooOop=********
  yellowNotepad=********
  `
	expected2 := `Account: other_account
Env Vars:
  superkilla=********
  superfreak=********
  frenchToast=********
  normcore=********
`
	live = ui.OutputWriter.String()
	if !strings.Contains(live, expected1) {
		t.Error("should have shankj3 block of env vars")
	}
	if !strings.Contains(live, expected2) {
		t.Error("should have other_account block of env vars")
	}
}

func testNew() *cmd {
	ui := cli.NewMockUi()
	c := &cmd{
		UI:     ui,
		config: commandhelper.NewTestClientConfig([]string{}),
	}
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.StringVar(&c.accountFilter, "account", "", "")
	return c
}

func BenchmarkOrganizeCred10(b *testing.B) {
	var credz []*pb.GenericCreds
	for n := 0; n < 10; n++ {
		credz = append(credz, creds...)
	}
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		organize(&pb.GenericWrap{Creds: credz})
	}
}

func BenchmarkOrganizeCred100(b *testing.B) {
	var credz []*pb.GenericCreds
	for n := 0; n < 100; n++ {
		credz = append(credz, creds...)
	}
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		organize(&pb.GenericWrap{Creds: credz})
	}
}

func organize2(cred *pb.GenericWrap) map[string][]*pb.GenericCreds {
	organizedCreds := make(map[string][]*pb.GenericCreds)
	for _, cred := range cred.Creds {
		acctCreds := organizedCreds[cred.AcctName]
		acctCreds = append(acctCreds, cred)
		organizedCreds[cred.AcctName] = acctCreds
	}
	return organizedCreds
}

func BenchmarkOrganize2Cred10(b *testing.B) {
	var credz []*pb.GenericCreds
	for n := 0; n < 10; n++ {
		credz = append(credz, creds...)
	}
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		organize2(&pb.GenericWrap{Creds: credz})
	}
}

func BenchmarkOrganize2Cred100(b *testing.B) {
	var credz []*pb.GenericCreds
	for n := 0; n < 100; n++ {
		credz = append(credz, creds...)
	}
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		organize2(&pb.GenericWrap{Creds: credz})
	}
}

var secret = "********"

var moreCreds = []*pb.GenericCreds{
	{
		AcctName:     "shankj3",
		Identifier:   "ENVIRONMENT",
		ClientSecret: secret,
		SubType:      pb.SubCredType_ENV,
	},
	{
		AcctName:     "shankj3",
		Identifier:   "EIINVIRONMENT",
		ClientSecret: secret,
		SubType:      pb.SubCredType_ENV,
	},
	{
		AcctName:     "shankj3",
		Identifier:   "whoopwhooOop",
		ClientSecret: secret,
		SubType:      pb.SubCredType_ENV,
	},
	{
		AcctName:     "shankj3",
		Identifier:   "yellowNotepad",
		ClientSecret: secret,
		SubType:      pb.SubCredType_ENV,
	},
	{
		AcctName:     "other_account",
		Identifier:   "superkilla",
		ClientSecret: secret,
		SubType:      pb.SubCredType_ENV,
	},
	{
		AcctName:     "other_account",
		Identifier:   "superfreak",
		ClientSecret: secret,
		SubType:      pb.SubCredType_ENV,
	},
	{
		AcctName:     "other_account",
		Identifier:   "frenchToast",
		ClientSecret: secret,
		SubType:      pb.SubCredType_ENV,
	},
	{
		AcctName:     "other_account",
		Identifier:   "normcore",
		ClientSecret: secret,
		SubType:      pb.SubCredType_ENV,
	},
}

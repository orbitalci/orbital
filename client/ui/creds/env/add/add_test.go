package envadd

import (
	"bytes"
	"testing"

	"github.com/go-test/deep"
	"github.com/mitchellh/cli"
	"github.com/shankj3/go-til/test"
	"github.com/level11consulting/ocelot/common/testutil"
	"github.com/level11consulting/ocelot/models/pb"

	"github.com/level11consulting/ocelot/client/commandhelper"
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
	if code := c.Run([]string{"-acct=jessishank", "ENVIRON1=MISSISSIPPI"}); code != 0 {
		t.Error("shouldn't fail")
	}
	expected := `The environment variable with the name ENVIRON1 already exists under the account jessishank. Do you wish to overwrite? Only a YES will continue and update the environment variable value, otherwise this entry will be skipped.`
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
	if code := c.Run([]string{"-acct=jessishank", "ENVIRON1=MISSISSIPPI"}); code != 0 {
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

func TestCmd_Run_badenv(t *testing.T) {
	ui := cli.NewMockUi()
	conf := commandhelper.NewTestClientConfig([]string{})
	c := &cmd{
		UI:     ui,
		config: conf,
	}
	c.init()
	client := c.config.Client.(*testutil.FakeGuideOcelotClient)
	client.Generics = &pb.GenericWrap{}
	if code := c.Run([]string{"-acct=jessishank", "envis:there"}); code != 1 {
		t.Error("should fail")
	}
	errmsg := ui.ErrorWriter.String()
	expected := "Bad environment variable. Must be in format ENV_VAR=ENV_VALUE \n"
	if errmsg != expected {
		t.Error(test.StrFormatErrors("err msg for bad env value", expected, errmsg))
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
	if code := c.Run([]string{"-acct=jessishank", "ENVIRON1=MISSISSIPPI", "ENV19=BEYONCE"}); code != 0 {
		t.Error("shouldn't fail")
	}
	expected := &pb.GenericWrap{
		Creds: []*pb.GenericCreds{
			{
				AcctName:     "jessishank",
				Identifier:   "ENVIRON1",
				ClientSecret: "MISSISSIPPI",
				SubType:      pb.SubCredType_ENV,
			},
			{
				AcctName:     "jessishank",
				Identifier:   "ENV19",
				ClientSecret: "BEYONCE",
				SubType:      pb.SubCredType_ENV,
			},
		},
	}
	if diff := deep.Equal(expected, client.Generics); diff != nil {
		t.Error(diff)
	}
}

func TestCmd_Run_nofile(t *testing.T) {
	ui := cli.NewMockUi()
	conf := commandhelper.NewTestClientConfig([]string{})
	c := &cmd{
		UI:     ui,
		config: conf,
	}
	c.init()
	client := c.config.Client.(*testutil.FakeGuideOcelotClient)
	client.Generics = &pb.GenericWrap{}
	if code := c.Run([]string{"-acct=jessishank", "-envfile=test-fixtures/notthere.yml"}); code != 1 {
		t.Error("should fail")
	}
	expected := `Error reading file at test-fixtures/notthere.yml, please check your filepath. 
Error is: 
    open test-fixtures/notthere.yml: no such file or directory
`
	errmsg := ui.ErrorWriter.String()
	if expected != errmsg {
		t.Error(test.StrFormatErrors("err msg", expected, errmsg))
	}
}

func TestCmd_Run_filepathbad(t *testing.T) {
	ui := cli.NewMockUi()
	conf := commandhelper.NewTestClientConfig([]string{})
	c := &cmd{
		UI:     ui,
		config: conf,
	}
	c.init()
	client := c.config.Client.(*testutil.FakeGuideOcelotClient)
	client.Generics = &pb.GenericWrap{}
	if code := c.Run([]string{"-acct=jessishank", "-envfile=test-fixtures/badenvs.yml"}); code != 1 {
		t.Error("should fail")
	}
	errmsg := ui.ErrorWriter.String()
	expected := `Could not unmarshal yaml file to a map of ENV_NAME: ENV_VALUE pairs. Please read the documentation and check your file at test-fixtures/badenvs.yml.
`
	if errmsg != expected {
		t.Error(test.StrFormatErrors("err msg", expected, errmsg))
	}
}

func TestCmd_Run_filepath(t *testing.T) {
	ui := cli.NewMockUi()
	conf := commandhelper.NewTestClientConfig([]string{})
	c := &cmd{
		UI:     ui,
		config: conf,
	}
	c.init()
	client := c.config.Client.(*testutil.FakeGuideOcelotClient)
	client.Generics = &pb.GenericWrap{}
	if code := c.Run([]string{"-acct=jessishank", "-envfile=test-fixtures/envs.yml"}); code != 0 {
		t.Error("shouldn't fail")
	}
	//expected := &pb.GenericWrap{
	//	Creds:[]*pb.GenericCreds{
	//		{
	//			AcctName: "jessishank",
	//			Identifier: "ENV1",
	//			ClientSecret: "TWELVE",
	//			SubType: pb.SubCredType_ENV,
	//		},
	//		{
	//			AcctName: "jessishank",
	//			Identifier: "DEPLOYKEY",
	//			ClientSecret: "SCMOOOOZLE",
	//			SubType: pb.SubCredType_ENV,
	//		},
	//		{
	//			AcctName: "jessishank",
	//			Identifier: "APIKEY",
	//			ClientSecret: "SPECIALBOiiii",
	//			SubType: pb.SubCredType_ENV,
	//		},
	//		{
	//			AcctName: "jessishank",
	//			Identifier: "nicenicecode",
	//			ClientSecret: "sweetbbknowswhatsup",
	//			SubType: pb.SubCredType_ENV,
	//		},
	//	},
	//}
	if len(client.Generics.Creds) != 4 {
		t.Error("not all creds uploaded")
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

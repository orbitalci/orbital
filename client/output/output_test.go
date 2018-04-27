package output

import (
	"context"
	"flag"
	"github.com/mitchellh/cli"
	"github.com/shankj3/go-til/test"
	"github.com/shankj3/ocelot/client/commandhelper"
	"strings"
	"testing"
)

func TestCmd_fromStorage(t *testing.T) {
	hash := "testinghash"
	streamText := "ayyyyayyyayyyayy\nwhyyywhyywhyywhyy\nnoonoonoonoo"
	ui := cli.NewMockUi()
	cliConf := commandhelper.NewTestClientConfig(strings.Split(streamText, "\n"))
	cmdd := &cmd{
		UI:        ui,
		config:    cliConf,
		OcyHelper: &commandhelper.OcyHelper{Hash: hash},
	}
	ctx := context.Background()
	exit := cmdd.fromStorage(ctx, hash)
	if exit != 0 {
		t.Error("non zero exit code")
	}
	text := ui.OutputWriter.String()
	if text != streamText+"\n" {
		test.StrFormatErrors("output", streamText+"\n", text)
	}
}

func TestCmd_RunMultipleBuilds(t *testing.T) {
	hash := "testinghash"
	ui := cli.NewMockUi()
	cliConf := commandhelper.NewTestClientConfig([]string{})
	cliConf.Theme = commandhelper.Default(true)
	cmdd := &cmd{
		UI:        ui,
		config:    cliConf,
		OcyHelper: &commandhelper.OcyHelper{Hash: hash},
	}
	cmdd.flags = flag.NewFlagSet("", flag.ContinueOnError)
	cmdd.flags.StringVar(&cmdd.OcyHelper.Hash, "hash", hash, "goal hash")
	var args []string
	exit := cmdd.Run(args)

	if exit != 0 {
		t.Error("non zero exit code")
	}

	expectedOutput := "it's your lucky day, there's 2 hashes matching that value. Please enter a more complete git hash"
	text := ui.OutputWriter.String()
	if !strings.HasPrefix(text, expectedOutput) {
		t.Error(test.StrFormatErrors("multiple hashes output", expectedOutput, text))
	}
}

//removing this because it requires mocking out grpc.ClientStream in our mocks, which is difficult

//func TestCmd_fromWerker(t *testing.T) {
//	var data = []struct{
//		hash string
//		stream string
//	}{
//		{"hashyhash", "al;ksdfjlksjfaslkdfj\n203948jfas;dkf8ewe\nalskdfjane8nxzlcfkue8@#$@#$@#$\n83nuadfn32"},
//		{"3jd8r32lks", "83242$#%@#%!#!@#!@\n)@!*$NASDFUEN\nfa;lskdjfal;ksdjf\nasdfasdfasdf"},
//		{"3jd8r232lks", "˚∂˜¨¨˙¬∂˚˜˜ππœ…“µß˙©¬˚˜˜…¬˚∆\n∂¬˚∆ƒ∂¬˚˜˜˜˜µ≤≈"},
//	}
//	for _, datum := range data {
//		lines := strings.Split(datum.stream, "\n")
//		buildRuntime := models.NewTestBuildRuntime(false, "", "", lines)
//		ui := cli.NewMockUi()
//		//cliConf := commandhelper.NewTestClientConfig(lines)
//		cmdd := &cmd{
//			UI: ui,
//			config: nil,
//			OcyHelper: &commandhelper.OcyHelper{Hash: datum.hash},
//		}
//		ctx := context.Background()
//		exit := cmdd.fromWerker(ctx, buildRuntime)
//		if exit != 0 {
//			t.Error("non zero exit code")
//		}
//		text := ui.OutputWriter.String()
//		if text != datum.stream + "\n" {
//			t.Error(test.StrFormatErrors("output", datum.stream + "\n", text))
//		}
//	}
//}

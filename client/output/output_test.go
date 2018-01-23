package output

import (
	"bitbucket.org/level11consulting/go-til/test"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/client/commandhelper"
	"context"
	"github.com/mitchellh/cli"
	"strings"
	"testing"
)

func TestCmd_fromStorage(t *testing.T) {
	hash := "testinghash"
	streamText := "ayyyyayyyayyyayy\nwhyyywhyywhyywhyy\nnoonoonoonoo"
	ui := cli.NewMockUi()
	cliConf := commandhelper.NewTestClientConfig(strings.Split(streamText, "\n"))
	cmdd := &cmd{
		UI: ui,
		config: cliConf,
		hash: hash,
	}
	ctx := context.Background()
	exit := cmdd.fromStorage(ctx, hash)
	if exit != 0 {
		t.Error("non zero exit code")
	}
	text := ui.OutputWriter.String()
	if text != streamText + "\n" {
		test.StrFormatErrors("output", streamText + "\n", text)
	}
}

func TestCmd_fromWerker(t *testing.T) {
	var data = []struct{
		hash string
		stream string
	}{
		{"hashyhash", "al;ksdfjlksjfaslkdfj\n203948jfas;dkf8ewe\nalskdfjane8nxzlcfkue8@#$@#$@#$\n83nuadfn32"},
		{"3jd8r32lks", "83242$#%@#%!#!@#!@\n)@!*$NASDFUEN\nfa;lskdjfal;ksdjf\nasdfasdfasdf"},
		{"3jd8r232lks", "˚∂˜¨¨˙¬∂˚˜˜ππœ…“µß˙©¬˚˜˜…¬˚∆\n∂¬˚∆ƒ∂¬˚˜˜˜˜µ≤≈"},
	}
	for _, datum := range data {
		lines := strings.Split(datum.stream, "\n")
		buildRuntime := models.NewTestBuildRuntime(false, "", "", lines)
		ui := cli.NewMockUi()
		//cliConf := commandhelper.NewTestClientConfig(lines)
		cmdd := &cmd{
			UI: ui,
			config: nil,
			hash: datum.hash,
		}
		ctx := context.Background()
		exit := cmdd.fromWerker(ctx, buildRuntime)
		if exit != 0 {
			t.Error("non zero exit code")
		}
		text := ui.OutputWriter.String()
		if text != datum.stream + "\n" {
			t.Error(test.StrFormatErrors("output", datum.stream + "\n", text))
		}
	}
}

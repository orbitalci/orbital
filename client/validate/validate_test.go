package validate

import (
	"testing"
	//"context"
	"github.com/mitchellh/cli"
	"flag"
	"strings"
)

func TestCmd_RunEmptyPath(t *testing.T) {
	ui := cli.NewMockUi()
	cmdd := &cmd{
		UI: ui,
	}
	cmdd.flags = flag.NewFlagSet("", flag.ContinueOnError)

	var args []string
	if exit := cmdd.Run(args); exit != -18511 {
		t.Error("should exit with code -18511 for help, exited with code ", exit)
	}

	//args = append(args, "/abc/def")

	expectedError := `

Usage: ocelot validate [filepath]
  Interacting with ocelot validator
  This client takes in an argument as a path to a local ocelot.yaml file
  Example: ocelot validate /home/mariannef/git/MyProject/ocelot.yml

`
	//if exit := cmdd.Run(args); exit != 1 {
	//	t.Error("should exit with error code 1", exit)
	//}

	errMsg := ui.OutputWriter.String()
	otherMsg := ui.ErrorWriter.String()
	if strings.Compare(expectedError, errMsg) != 0 {
		t.Errorf("output and expected not the same,  \n" +
			"expected:\n%s\ngot:\n%s", "an error", errMsg)
	}
	if strings.Compare(expectedError, otherMsg) != 0 {
		t.Errorf("output and expected not the same,  \n" +
			"expected:\n%s\ngot:\n%s", "an error", errMsg)
	}
}
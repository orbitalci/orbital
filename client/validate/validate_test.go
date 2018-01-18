package validate

import (
	"testing"
	//"context"
	"github.com/mitchellh/cli"
	"flag"
	"strings"
	"os"
)

func TestCmd_RunPathNoFile(t *testing.T) {
	ui := cli.NewMockUi()
	cmdd := &cmd{
		UI: ui,
	}
	cmdd.flags = flag.NewFlagSet("", flag.ContinueOnError)

	args := []string{"/abc/def/test"}

	expectedError := `Could not read file at /abc/def/test
Error: open /abc/def/test: no such file or directory
`
	if exit := cmdd.Run(args); exit != 1 {
		t.Error("should exit with error code 1", exit)
	}

	errMsg := ui.ErrorWriter.String()
	if strings.Compare(expectedError, errMsg) != 0 {
		t.Errorf("output and expected not the same,  \n" +
			"expected:\n%s\ngot:\n%s", expectedError, errMsg)
	}
}

func TestCmd_RunPathFileNoProcess(t *testing.T) {
	ui := cli.NewMockUi()
	pwd, _ := os.Getwd()
	cmdd := &cmd{
		UI: ui,
	}
	cmdd.flags = flag.NewFlagSet("", flag.ContinueOnError)

	fileNoExist := []string{pwd + "/test-fixtures/wrong-ocelot.yml"}

	expectedError := `Could not process file, please check make sure the file at /Users/mariannefeng/go/src/bitbucket.org/level11consulting/ocelot/client/validate/test-fixtures/wrong-ocelot.yml exists
Error: yaml: unmarshal errors:
  line 1: cannot unmarshal !!str ` + "`wrong`" + ` into protos.BuildConfig
`
	if exit := cmdd.Run(fileNoExist); exit != 1 {
		t.Error("should exit with error code 1", exit)
	}

	errMsg := ui.ErrorWriter.String()
	if strings.Compare(expectedError, errMsg) != 0 {
		t.Errorf("output and expected not the same,  \n" +
			"expected:\n%s\ngot:\n%s", expectedError, errMsg)
	}
}

func TestCmd_RunPathFileName(t *testing.T) {
	ui := cli.NewMockUi()
	pwd, _ := os.Getwd()
	cmdd := &cmd{
		UI: ui,
	}
	cmdd.flags = flag.NewFlagSet("", flag.ContinueOnError)

	badName := []string{pwd + "/test-fixtures/bad-name.yml"}

	expectedError := `Your file must be named ocelot.yml
`
	if exit := cmdd.Run(badName); exit != 1 {
		t.Error("should exit with error code 1", exit)
	}

	errMsg := ui.ErrorWriter.String()
	if strings.Compare(expectedError, errMsg) != 0 {
		t.Errorf("output and expected not the same,  \n" +
			"expected:\n%s\ngot:\n%s", expectedError, errMsg)
	}
}

func TestCmd_RunPathFileWrongFormat(t *testing.T) {
	ui := cli.NewMockUi()
	pwd, _ := os.Getwd()
	cmdd := &cmd{
		UI: ui,
	}
	cmdd.flags = flag.NewFlagSet("", flag.ContinueOnError)

	badName := []string{pwd + "/test-fixtures/ocelot.yml"}

	expectedError := `Invalid ocelot.yml file: BuildTool must be specified
`
	if exit := cmdd.Run(badName); exit != 1 {
		t.Error("should exit with error code 1", exit)
	}

	errMsg := ui.ErrorWriter.String()
	if strings.Compare(expectedError, errMsg) != 0 {
		t.Errorf("output and expected not the same,  \n" +
			"expected:\n%s\ngot:\n%s", expectedError, errMsg)
	}
}


//TODO: why can't I read the output from help message? Is it cause I'm not explicitly calling it?
func TestCmd_RunEmptyPath(t *testing.T) {
//	ui := cli.NewMockUi()
//	cmdd := &cmd{
//		UI: ui,
//	}
//	cmdd.flags = flag.NewFlagSet("", flag.ContinueOnError)
//
//	var args []string
//	if exit := cmdd.Run(args); exit != -18511 {
//		t.Error("should exit with code -18511 for help, exited with code ", exit)
//	}
//
//	expectedError := `
//
//Usage: ocelot validate [filepath]
//  Interacting with ocelot validator
//  This client takes in an argument as a path to a local ocelot.yaml file
//  Example: ocelot validate /home/mariannef/git/MyProject/ocelot.yml
//
//`
//
//	errMsg := ui.OutputWriter.String()
//	if strings.Compare(expectedError, errMsg) != 0 {
//		t.Errorf("output and expected not the same,  \n" +
//			"expected:\n%s\ngot:\n%s", "an error", errMsg)
//	}
}


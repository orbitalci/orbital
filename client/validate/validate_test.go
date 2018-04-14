package validate

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"bitbucket.org/level11consulting/ocelot/common/testutil"
	"github.com/mitchellh/cli"
)

func TestCmd_RunPathNoFile(t *testing.T) {
	ui := cli.NewMockUi()
	cmdd := &cmd{
		UI: ui,
		ocelotFileLoc: "/abc/def/test",
	}
	cmdd.flags = flag.NewFlagSet("", flag.ContinueOnError)
	var args []string
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
	testutil.BuildServerHack(t)
	ui := cli.NewMockUi()
	pwd, _ := os.Getwd()
	cmdd := &cmd{
		UI: ui,
		ocelotFileLoc: pwd + "/test-fixtures/wrong-ocelot.yml",
	}
	cmdd.flags = flag.NewFlagSet("", flag.ContinueOnError)

	var args []string
	filepth := os.ExpandEnv("$HOME/go/src/bitbucket.org/level11consulting/ocelot/client/validate/test-fixtures/wrong-ocelot.yml")
	expectedError := fmt.Sprintf(`Could not process file, please check make sure the file at %s exists
Error: yaml: unmarshal errors:
  line 1: cannot unmarshal !!str ` + "`wrong`" + ` into protos.BuildConfig
`, filepth)
	if exit := cmdd.Run(args); exit != 1 {
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
		ocelotFileLoc: pwd + "/test-fixtures/bad-name.yml",
	}
	cmdd.flags = flag.NewFlagSet("", flag.ContinueOnError)

	var args []string
	expectedError := `Your file must be named ocelot.yml
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

func TestCmd_RunPathFileWrongFormat(t *testing.T) {
	ui := cli.NewMockUi()
	pwd, _ := os.Getwd()
	cmdd := &cmd{
		UI: ui,
		ocelotFileLoc: pwd + "/test-fixtures/ocelot.yml",
	}
	cmdd.flags = flag.NewFlagSet("", flag.ContinueOnError)

	var args []string
	expectedError := `Invalid ocelot.yml file: BuildTool must be specified
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
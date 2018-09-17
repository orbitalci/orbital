package validate

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/mitchellh/cli"
	"github.com/shankj3/ocelot/common/testutil"
)

func TestCmd_RunPathNoFile(t *testing.T) {
	ui := cli.NewMockUi()
	cmdd := &cmd{
		UI:            ui,
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
		t.Errorf("output and expected not the same,  \n"+
			"expected:\n%s\ngot:\n%s", expectedError, errMsg)
	}
}

func TestCmd_RunPathFileNoProcess(t *testing.T) {
	testutil.BuildServerHack(t)
	ui := cli.NewMockUi()
	pwd, _ := os.Getwd()
	cmdd := &cmd{
		UI:            ui,
		ocelotFileLoc: pwd + "/test-fixtures/wrong-ocelot.yml",
	}
	cmdd.flags = flag.NewFlagSet("", flag.ContinueOnError)

	var args []string
	filepth := os.ExpandEnv("$HOME/go/src/github.com/shankj3/ocelot/client/validate/test-fixtures/wrong-ocelot.yml")
	expectedError := fmt.Sprintf(`Could not process file, please check make sure the file at %s exists
Error: yaml: unmarshal errors:
  line 1: cannot unmarshal !!str `+"`wrong`"+` into pb.BuildConfig
`, filepth)
	if exit := cmdd.Run(args); exit != 1 {
		t.Error("should exit with error code 1", exit)
	}

	errMsg := ui.ErrorWriter.String()
	if strings.Compare(expectedError, errMsg) != 0 {
		t.Errorf("output and expected not the same,  \n"+
			"expected:\n%s\ngot:\n%s", expectedError, errMsg)
	}
}

func TestCmd_RunPathFileName(t *testing.T) {
	ui := cli.NewMockUi()
	pwd, _ := os.Getwd()
	cmdd := &cmd{
		UI:            ui,
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
		t.Errorf("output and expected not the same,  \n"+
			"expected:\n%s\ngot:\n%s", expectedError, errMsg)
	}
}

func TestCmd_RunPathFileWrongFormat(t *testing.T) {
	ui := cli.NewMockUi()
	pwd, _ := os.Getwd()
	cmdd := &cmd{
		UI:            ui,
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
		t.Errorf("output and expected not the same,  \n"+
			"expected:\n%s\ngot:\n%s", expectedError, errMsg)
	}
}

func TestCmd_Run_nofilepathloc(t *testing.T) {
	ui := cli.NewMockUi()
	//pwd, _ := os.Getwd()
	cmdd := New(ui)
	var args = []string{"test-fixtures/valid/ocelot.yml"}
	if exit := cmdd.Run(args); exit != 0 {
		t.Log(string(ui.OutputWriter.Bytes()))
		t.Log(string(ui.ErrorWriter.Bytes()))
		t.Error("should exit with error code 0: ", exit)
	}
}

func TestCmd_Run_validateWithBranch(t *testing.T) {
	ui := cli.NewMockUi()
	//pwd, _ := os.Getwd()
	cmdd := New(ui)
	var args = []string{"-branch=develop", "test-fixtures/valid/ocelot.yml"}
	if exit := cmdd.Run(args); exit != 1 {
		t.Error("should exit with error code 1: ", exit)
	}
	errr := string(ui.ErrorWriter.Bytes())
	expectedErr := "This branch would not build, the validation error was: branch develop not in the acceptable branches list: master\n"
	if diff := deep.Equal(expectedErr, errr); diff != nil {
		t.Error(diff)
	}
}

func TestCmd_Run_fromCurrentDirectory(t *testing.T) {
	ui := cli.NewMockUi()
	pwd, _ := os.Getwd()
	defer os.Chdir(pwd)
	cmdd := New(ui)
	os.Chdir(filepath.Join("test-fixtures", "valid"))
	var args []string
	if exit := cmdd.Run(args); exit != 0 {
		t.Error("should exit with 0: ", exit)
	}
	t.Log(string(ui.OutputWriter.Bytes()))
	t.Log(string(ui.ErrorWriter.Bytes()))
}

func TestCmd_Run_fromCurrentDirectory_notFound(t *testing.T) {
	ui := cli.NewMockUi()
	cmdd := New(ui)
	var args []string
	if exit := cmdd.Run(args); exit != 1 {
		t.Error("should exit with 1: ", exit)
	}
	errorStr := string(ui.ErrorWriter.Bytes())
	expected := "Could not find ocelot.yml in current directory. Please pass ocelot.yml location as an argument.\n"
	if diff := deep.Equal(expected, errorStr); diff != nil {
		t.Errorf("unexpected error output: %#v", diff)
	}
}

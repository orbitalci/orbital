package ocyinit

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/mitchellh/cli"
	"github.com/shankj3/go-til/test"
)

func TestNewAndOtherStaticStuff(t *testing.T) {
	ui := &cli.MockUi{}
	init := New(ui)
	t.Log(init.flags)
	if notify := init.flags.Lookup("notify"); notify == nil {
		t.Error("notify shoudl be in flags")
	}
	if renderTag := init.flags.Lookup("render-tag"); renderTag == nil {
		t.Error("render-tag should be in flags")
	}
	if init.GetClient() == nil || init.GetClient() != init.config.Client {
		t.Error("client should be set")
	}
	if init.GetUI() == nil || init.GetUI() != init.UI {
		t.Error("ui should be set")
	}
	if init.GetConfig() == nil {
		t.Error("config should be set")
	}
	if init.Synopsis() != "create a skeleton ocelot.yml file" {
		t.Error(test.StrFormatErrors("synopsis", "create a skeleton ocelot.yml file", init.Synopsis()))
	}
	if init.Help() != help {
		t.Error(test.StrFormatErrors("help", help, init.Help()))
	}
}

func getNewUI(t *testing.T) (*cmd, *cli.MockUi, func()) {
	t.Log(t.Name())
	ui := &cli.MockUi{}
	init := New(ui)
	path := filepath.Join("test-fixtures", t.Name())
	os.MkdirAll(path, 0755)
	cur, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
		return nil, ui, func() {}
	}
	err = os.Chdir(path)
	if err != nil {
		t.Fatal(err)
		return nil, ui, func() {}
	}
	return init, ui, func() {
		os.Chdir(cur)
		os.RemoveAll(path)
	}
}

// happy path
func TestCmd_Run(t *testing.T) {
	init, ui, clean := getNewUI(t)
	args := strings.Split("-notify -render-tag", " ")
	defer clean()
	code := init.Run(args)
	defer os.Remove("ocelot.yml")
	if code != 0 {
		t.Log(string(ui.OutputWriter.Bytes()))
		t.Log(string(ui.ErrorWriter.Bytes()))
		t.Error("should have exited 0")
	}
	//check for ocelot.yml
	ocyfile, err := ioutil.ReadFile("ocelot.yml")
	if err != nil {
		t.Error(err)
		return
	}
	expecetd := fmt.Sprintf(skeleton, "machineTag", notifySkeleton)
	if string(ocyfile) != expecetd {
		t.Error(test.StrFormatErrors("rendered ocelot.yml", expecetd, string(ocyfile)))
	}
}

func TestCmd_Run_fileexists(t *testing.T) {
	init, ui, clean := getNewUI(t)
	var args []string
	defer clean()
	d1 := []byte("ocelottttt")
	err := ioutil.WriteFile("ocelot.yml", d1, 0600)
	if err != nil {
		t.Error(err)
		return
	}
	code := init.Run(args)
	defer os.Remove("ocelot.yml")
	if code != 1 {
		t.Error("should have failed?")
	}
	// have to take out newline because errorwriter add it
	live := string(ui.ErrorWriter.Bytes())
	expected := "There is already an ocelot.yml file in this directory, and I'm not going to overwrite it. Please delete the file if you wish to continue with generating a skeleton.\n"
	if diff := deep.Equal(live, expected); diff != nil {
		t.Error(diff)
	}

}

func TestCmd_Run_nopermissions(t *testing.T) {
	init, ui, clean := getNewUI(t)
	var args []string
	defer clean()
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
		return
	}
	testDir := filepath.Join("test-fixtures", t.Name())
	initDir := filepath.Dir(filepath.Dir(wd))
	os.Chdir(initDir)
	t.Log(os.Getwd())
	// change permissions of test directory to read/execute
	err = os.Chmod(testDir, 0500)
	if err != nil {
		t.Error(err)
		return
	}
	err = os.Chdir(testDir)
	if err != nil {
		t.Error(err)
		return
	}
	code := init.Run(args)
	if code == 0 {
		t.Error("should have failed")
	}
	errorStr := string(ui.ErrorWriter.Bytes())
	outStr := string(ui.OutputWriter.Bytes())
	if diff := deep.Equal(errorStr, "Unable to write file to ocelot.yml in this location, will print to stdout instead. For your edification, the error is: open ocelot.yml: permission denied\n"); diff != nil {
		t.Error(diff)
	}
	expected := fmt.Sprintf(skeleton, "image", "") + "\n"
	if diff := deep.Equal(expected, outStr); diff != nil {
		t.Error(diff)
	}
}

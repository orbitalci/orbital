package exec

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	//"time"

	"github.com/go-test/deep"
	"github.com/google/uuid"
	"github.com/shankj3/go-til/test"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/build/basher"
	"github.com/shankj3/ocelot/common/testutil"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
)

func setupExec(t *testing.T) *Exec {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	ocyPrefix := filepath.Join(dir, "test-fixtures/exec")
	bshr, err := basher.NewBasher("", "", "localhost", ocyPrefix)
	if err != nil {
		t.Fatal(err)
	}
	facts := &models.WerkerFacts{
		Uuid:        uuid.New(),
		WerkerType:  models.Exec,
		LoopbackIp:  "localhost",
		RegisterIP:  "localhost",
		ServicePort: "9999",
		Dev:         true,
	}
	return NewExecBuilder(bshr, facts).(*Exec)

}

func TestExec_Setup(t *testing.T) {
	hash := "execbuild123"
	exc := setupExec(t)
	rmTar := testutil.TarTemplates(t, "../builder/exec/test-fixtures/werker_files.tar", "../../template/")
	defer rmTar(t)
	go testutil.CreateDoThingsWebServer("./test-fixtures/werker_files.tar", "9999")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logout := make(chan []byte, 1000)
	idChan := make(chan string, 1)
	task := &pb.WerkerTask{CheckoutHash: hash}
	res, id := exc.Setup(ctx, logout, idChan, task, nil, "9999")
	close(logout)
	defer os.RemoveAll("./test-fixtures/exec/.ocelot")
	if res.Status == pb.StageResultVal_FAIL {
		for i := range logout {
			fmt.Println(string(i))
		}
		t.Log(res.Messages)
		t.Log(res.Error)
		t.Error("setup should have passed")
	}
	if id != hash {
		t.Error(test.StrFormatErrors("build 'docker' id", hash, id))
	}
	untarredFiles, err := ioutil.ReadDir("./test-fixtures/exec/.ocelot")
	if err != nil {
		t.Error(err)
		return
	}
	fileMap := createFileMap(untarredFiles)
	if _, ok := fileMap["bb_download.sh"]; !ok {
		t.Log(fileMap)
		t.Error("bb_download should have been untarred and put in .ocelot")
	}
	if _, ok := fileMap["get_ssh_key.sh"]; !ok {
		t.Log(fileMap)
		t.Error("get_ssh_key should have been untarred and put in .ocelot")
	}
	if _, ok := fileMap["install_deps.sh"]; !ok {
		t.Log(fileMap)
		t.Error("install_deps.sh shoudl have been untarred and put in .ocelot")
	}
	DIREC, _ := os.Getwd()
	fmt.Println("WORKIN DIRECTORY!", DIREC)

}

func TestExec_Setupfail(t *testing.T) {
	hash := "execbuild123"
	exc := setupExec(t)
	// webserver isn't running, should fail to connect
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logout := make(chan []byte, 1000)
	idChan := make(chan string, 1)
	task := &pb.WerkerTask{CheckoutHash: hash}
	res, _ := exc.Setup(ctx, logout, idChan, task, nil, "9999")
	close(logout)
	var out string
	for i := range logout {
		out += string(i) + "\n"
	}
	if res.Status != pb.StageResultVal_FAIL {
		t.Error("should fail as there is no web server for downloading templates")
	}
	if diff := deep.Equal(res.Messages, []string{"failed to download templates ✗"}); diff != nil {
		t.Error(diff)
	}
}

func TestExec_Execute(t *testing.T) {
	DIREC, _ := os.Getwd()
	t.Log("WORKIN DIRECTORY 0!", DIREC)
	ctx := context.Background()
	hash := "execbuild123"
	hashDir := "./test-fixtures/exec/.ocelot/" + hash
	err := os.MkdirAll(hashDir, 0700)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(hashDir, "README.md"), []byte("glorious readme"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	exc := setupExec(t)
	runStage := &pb.Stage{
		Env:    []string{"HEREBEMYTESTVAR=3"},
		Script: []string{"echo $HEREBEMYTESTVAR", "ls"},
		Name:   "executethisbish",
	}
	var logs string
	logsdone := make(chan int, 1)
	logout := make(chan []byte)
	go func() {
		for i := range logout {
			logs += string(i) + "\n"
			fmt.Println(logs)
		}
		close(logsdone)
	}()
	DIREC, _ = os.Getwd()
	t.Log("WORKIN DIRECTORY 1!", DIREC)
	res := exc.Execute(ctx, runStage, logout, hash)
	DIREC, _ = os.Getwd()
	t.Log("WORKIN DIRECTORY 2!", DIREC)
	close(logout)
	<-logsdone
	if res.Status == pb.StageResultVal_FAIL {
		t.Error("should have passed, error is:" + res.Error)
		return
	}
	expected := `EXECUTETHISBISH | 3
EXECUTETHISBISH | README.md
`
	if logs != expected {
		t.Error(test.StrFormatErrors("output", expected, logs))
	}
}

func createFileMap(files []os.FileInfo) map[string]int {
	fm := make(map[string]int)
	for _, file := range files {
		fm[file.Name()] = 1
	}
	return fm
}

func Test_prepCmds(t *testing.T) {
	var tests = []struct {
		name     string
		cmds     []string
		expected [3]string
	}{
		{"already parsed", []string{"/bin/bash", "-c", "echo ayyyyyyy"}, [3]string{"/bin/bash", "-c", "echo ayyyyyyy"}},
		{"already parsed sh", []string{"/bin/sh", "-c", "echo ayyyyyyy"}, [3]string{"/bin/sh", "-c", "echo ayyyyyyy"}},
		{"jumbled", []string{"echo 'halleloo'", "ls -latr", "tail -100f /var/log/system.log"}, [3]string{"/bin/bash", "-c", "echo 'halleloo' && ls -latr && tail -100f /var/log/system.log"}},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			prepped := prepCmds(tst.cmds)
			if diff := deep.Equal(prepped, tst.expected); diff != nil {
				t.Error(diff)
			}
		})
	}
	DIREC, _ := os.Getwd()
	fmt.Println("WORKIN DIRECTORY!", DIREC)
}

func TestExec_execute(t *testing.T) {
	DIREC, _ := os.Getwd()
	t.Log("WORKIN DIRECTORY!", DIREC)
	exc := setupExec(t)
	ctx := context.Background()
	su := build.InitStageUtil("execccctessst")
	logout := make(chan []byte)
	logsdone := make(chan bool, 1)
	var logs string
	go func() {
		for i := range logout {
			logs += string(i) + "\n"
		}
		close(logsdone)
	}()
	DIREC, _ = os.Getwd()
	t.Log("WORKIN DIRECTORY!", DIREC)
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	path := filepath.Join(dir, "test-fixtures")
	res := exc.execute(ctx, su, []string{"PRIVTEST=execute"}, []string{"cd " + path, "echo $(pwd)"}, logout)
	close(logout)
	<-logsdone
	t.Log(logs)
	if res.Status == pb.StageResultVal_FAIL {
		t.Error("stage should not have failed, error is: " + res.Error)
	}
	if !strings.Contains(logs, "builder/exec/test-fixtures") {
		t.Error("pwd should return the test-fixtures direc, output is: \n" + logs)
	}
}

func TestExec_executeFail(t *testing.T) {
	DIREC, _ := os.Getwd()
	t.Log("WORKIN DIRECTORY!", DIREC)
	exc := setupExec(t)
	ctx := context.Background()
	su := build.InitStageUtil("execccctessst")
	logout := make(chan []byte)
	logsdone := make(chan bool, 1)
	var logs string
	go func() {
		for i := range logout {
			logs += string(i) + "\n"
		}
		close(logsdone)
	}()
	DIREC, _ = os.Getwd()
	t.Log("WORKIN DIRECTORY!", DIREC)
	stage := &pb.Stage{
		Name:   "execccctessst",
		Script: []string{"echo hi | grep onomotopoeia"},
		Env:    []string{"PRIVTEST=execute"},
	}
	res := exc.ExecuteIntegration(ctx, stage, su, logout)
	close(logout)
	<-logsdone
	t.Log(logs)
	if res.Status != pb.StageResultVal_FAIL {
		t.Error("stage should have failed, error is: " + res.Error)
	}
}

func TestExec_SetGlobalEnv(t *testing.T) {
	exc := setupExec(t)
	ctx := context.Background()
	su := build.InitStageUtil("globalenvtesssst")
	exc.SetGlobalEnv([]string{"ONETWOTHREE=FOURFIVESIX", "THINK=GLOBALLY"})
	logout := make(chan []byte)
	logsdone := make(chan bool, 1)
	var logs string
	go func() {
		for i := range logout {
			logs += string(i) + "\n"
		}
		close(logsdone)
	}()
	res := exc.execute(ctx, su, []string{"PRIVTEST=execute"}, []string{"echo $THINK", "echo $ONETWOTHREE", "echo $PRIVTEST"}, logout)
	close(logout)
	<-logsdone
	if res.Status == pb.StageResultVal_FAIL {
		t.Error("stage should not have failed, error is: " + res.Error)
	}
	expected := `GLOBALENVTESSSST | GLOBALLY
GLOBALENVTESSSST | FOURFIVESIX
GLOBALENVTESSSST | execute
`
	if expected != logs {
		t.Error(test.StrFormatErrors("log output", expected, logs))
	}
	DIREC, _ := os.Getwd()
	t.Log("WORKIN DIRECTORY!", DIREC)

}

func TestExec_Init(t *testing.T) {
	exc := setupExec(t)
	ctx := context.Background()
	logout := make(chan []byte)
	res := exc.Init(ctx, "hash", logout)
	if res.Status != pb.StageResultVal_PASS {
		t.Error("should pass as nothing is happening")
	}
	if res.Stage != "INIT" {
		t.Error("stage name should be init")
	}
}

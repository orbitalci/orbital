package ssh

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"testing"

	"github.com/go-test/deep"
	"github.com/shankj3/go-til/test"
	"github.com/level11consulting/orbitalci/build"
	"github.com/level11consulting/orbitalci/build/basher"
	"github.com/level11consulting/orbitalci/models"
	"github.com/level11consulting/orbitalci/models/pb"
)

func TestSSH_Initfail(t *testing.T) {
	sshfacts := &models.SSHFacts{
		User:  "root",
		Host:  "localhost",
		Port:  1234,
		KeyFP: "id",
	}
	werkerFacts := &models.WerkerFacts{
		Uuid:        uuid.New(),
		WerkerType:  models.SSH,
		LoopbackIp:  "docker.for.mac.localhost",
		RegisterIP:  "docker.for.mac.localhost",
		ServicePort: "",
		Dev:         true,
		Ssh:         sshfacts,
	}
	bshr, err := basher.NewBasher("", "", "docker.for.mac.localhost", "/tmp")
	ssh, err := NewSSHBuilder(bshr, werkerFacts)
	if err != nil {
		t.Fatal(err)
	}
	logout := make(chan []byte, 1000)
	res := ssh.Init(context.Background(), "hashhash", logout)
	close(logout)
	if res.Status != pb.StageResultVal_FAIL {
		t.Error("should fial, there is no ssh container up")
		return
	}
	if diff := deep.Equal(res.Messages, []string{"Failed to initialize ssh builder " + models.FAILED}); diff != nil {
		t.Error(diff)
	}
}

func TestSSH_GetContainerId(t *testing.T) {
	sshfacts := &models.SSHFacts{
		User:  "root",
		Host:  "localhost",
		Port:  1234,
		KeyFP: "id",
	}
	werkerFacts := &models.WerkerFacts{
		Uuid:        uuid.New(),
		WerkerType:  models.SSH,
		LoopbackIp:  "docker.for.mac.localhost",
		RegisterIP:  "docker.for.mac.localhost",
		ServicePort: "",
		Dev:         true,
		Ssh:         sshfacts,
	}
	bshr, err := basher.NewBasher("", "", "docker.for.mac.localhost", "/tmp")
	ssh, err := NewSSHBuilder(bshr, werkerFacts)
	if err != nil {
		t.Fatal(err)
	}
	if ssh.GetContainerId() != "" {
		t.Error("containerId isn't a managed field by the ssh struct? why is this set")
	}
}

// runs all the tests using one ssh container
func TestSSH(t *testing.T) {
	task := &pb.WerkerTask{CheckoutHash: "TESTHASHAYYY"}
	ssher, ctx, cancel, tarRm, cleaner := SetupSSHBuilder(t, 2222, "3833")
	defer tarRm(t)
	defer cleaner()
	defer cancel()
	defer ssher.Close()
	t.Run("SSH_setup", func(t *testing.T) {
		testSSH_Setup(t, ssher, ctx, task)
	})
	t.Run("SSH_setup2", func(t *testing.T) {
		testSSH_Setup2(t, ssher, ctx, task)
	})
	t.Run("execute", func(t *testing.T) {
		testSSH_execute(t, ssher, ctx)
	})
	t.Run("execute integration", func(t *testing.T) {
		testSSH_ExecuteIntegration(t, ssher, ctx)
	})
	t.Run("execute 2", func(t *testing.T) {
		testSSH_Execute(t, ssher, ctx, task.CheckoutHash)
	})
}

func testSSH_Setup(t *testing.T, ssher build.Builder, ctx context.Context, task *pb.WerkerTask) {
	logt := make(chan []byte, 100)
	logdone := make(chan int)
	var output string
	go func() {
		for i := range logt {
			output += string(i) + "\n"
		}
		close(logdone)
	}()
	res := ssher.Init(ctx, task.CheckoutHash, logt)
	close(logt)
	<-logdone
	if res.Status != pb.StageResultVal_PASS {
		t.Log(output)
		t.Error("should pass, error is :" + res.Error)
		return
	}
	expected := "Successfully established ssh connection " + models.CHECKMARK
	if res.Messages[0] != expected {
		t.Error(test.StrFormatErrors("init messages", expected, res.Messages[0]))
		return
	}
	logout := make(chan []byte, 1000)
	defer close(logout)
	dockerId := make(chan string, 1)
	result, _ := ssher.Setup(ctx, logout, dockerId, task, nil, "")
	expectedMsgs := []string{"successfully downloaded templates " + models.CHECKMARK}
	if diff := deep.Equal(result.Messages, expectedMsgs); diff != nil {
		t.Error(diff)
	}
}

// failure scenario
func testSSH_Setup2(t *testing.T, ssher build.Builder, ctx context.Context, task *pb.WerkerTask) {
	logt := make(chan []byte)
	logdone := make(chan int)
	var output string
	go func() {
		for i := range logt {
			output += string(i) + "\n"
		}
		close(logdone)
	}()
	res := ssher.Init(ctx, task.CheckoutHash, logt)
	close(logt)
	<-logdone
	if res.Status != pb.StageResultVal_PASS {
		t.Log(res.Messages)
		t.Log(output)
		t.Error("should pass, error is: " + res.Error)
		return
	}
	expected := "Successfully established ssh connection " + models.CHECKMARK
	if res.Messages[0] != expected {
		t.Error(test.StrFormatErrors("init messages", expected, res.Messages[0]))
		return
	}
	logout := make(chan []byte, 1000)
	defer close(logout)
	dockerId := make(chan string, 1)
	sher := ssher.(*SSH)
	sher.ServicePort = "8235"
	result, _ := sher.Setup(ctx, logout, dockerId, task, nil, "")
	if result.Status != pb.StageResultVal_FAIL {
		t.Error("should have failed")
	}
	if diff := deep.Equal(result.Messages, []string{"failed to download templates " + models.FAILED}); diff != nil {
		t.Error(diff)
	}
}

func testSSH_execute(t *testing.T, sher build.Builder, ctx context.Context) {
	ssher := sher.(*SSH)
	ssher.SetGlobalEnv([]string{"GLOBALENV=1234567"})
	logout := make(chan []byte)
	logdone := make(chan bool, 1)
	var live string
	go func() {
		for i := range logout {
			live += string(i) + "\n"
		}
		close(logdone)
	}()
	res := ssher.execute(ctx, build.InitStageUtil("SSHEXEC"), []string{"AYYY=123"}, []string{"echo 'SUPERCALLAFRAGILISTICEXPIALADOCIOUS'"}, logout)
	close(logout)
	<-logdone
	if res.Status == pb.StageResultVal_FAIL {
		t.Log(res.Error)
		t.Log(res.Messages)
		t.Error("test should have passed")
	}
	if live != "SSHEXEC | SUPERCALLAFRAGILISTICEXPIALADOCIOUS\n" {
		t.Error(test.StrFormatErrors("log output", "SSHEXEC | SUPERCALLAFRAGILISTICEXPIALADOCIOUS\n", live))
	}
}

func testSSH_ExecuteIntegration(t *testing.T, ssher build.Builder, ctx context.Context) {
	ssher.SetGlobalEnv([]string{"GLOBALENV=1234567"})
	logout := make(chan []byte)
	logdone := make(chan bool, 1)
	var live string
	go func() {
		for i := range logout {
			live += string(i) + "\n"
		}
		close(logdone)
	}()
	stage := &pb.Stage{
		Env:    []string{"AYYY=123"},
		Script: []string{"echo 'SUPERCALLAFRAGILISTICEXPIALADOCIOUS' | grep banana"},
		Name:   "SSHEXEC",
	}
	res := ssher.ExecuteIntegration(ctx, stage, build.InitStageUtil(stage.Name), logout)
	if res.Status == pb.StageResultVal_PASS {
		t.Error("exec should not have passed, there is no banana in the echo!")
	}
	if diff := deep.Equal(res.Messages, []string{"failed to complete sshexec stage ✗"}); diff != nil {
		t.Error(diff)
	}
}

func testSSH_Execute(t *testing.T, sher build.Builder, ctx context.Context, hash string) {
	logout := make(chan []byte, 1000)
	ssher := sher.(*SSH)
	res := ssher.execute(ctx, build.InitStageUtil("execute"), []string{}, []string{"mkdir -p /tmp/.ocelot/" + hash, fmt.Sprintf("touch /tmp/.ocelot/%s/README.md", hash)}, logout)
	close(logout)
	if res.Status == pb.StageResultVal_FAIL {
		var out string
		for i := range logout {
			out += string(i) + "\n"
		}
		t.Error("unable to set up hash direc for test, output is: \n" + out)
	}
	//Execute(ctx context.Context, actions *pb.Stage, logout chan []byte, commitHash string) *pb.Result
	execStage := &pb.Stage{
		Script: []string{"echo 'ayyyyyy'", "ls"},
		Name:   "testcd",
	}
	var out string
	logout = make(chan []byte)
	logsdone := make(chan bool, 1)
	go func() {
		for i := range logout {
			out += string(i) + "\n"
		}
		close(logsdone)
	}()
	res = ssher.Execute(ctx, execStage, logout, hash)
	close(logout)
	<-logsdone
	if res.Status == pb.StageResultVal_FAIL {
		t.Error("exec stage should have passed")
	}
	expected := `TESTCD | ayyyyyy
TESTCD | README.md
`
	if expected != out {
		t.Error(test.StrFormatErrors("log output", expected, out))
	}
}

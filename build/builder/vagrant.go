package builder
/*
vagrant implementation of builder should:
  SETUP: create directory for vagrantfile, run vagrant up
  EXECUTION: use crypto/ssh library to
*/

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"bitbucket.org/level11consulting/ocelot/build"
	"bitbucket.org/level11consulting/ocelot/build/basher"
	cred "bitbucket.org/level11consulting/ocelot/common/credentials"
	"bitbucket.org/level11consulting/ocelot/models"
	"bitbucket.org/level11consulting/ocelot/models/pb"
)


type Vagrant struct {
	globalEnvs     []string
	*basher.Basher
}


// createVagrantDirec will create a directory for Vagrantfiles in the ocelot directory under /vagrant/<hash>
func (v *Vagrant) createVagrantDirec(hash string) error {
	err := os.MkdirAll(v.getVagrantDirec(hash), os.FileMode(0700))
	return err
}

func (v *Vagrant) getVagrantDirec(hash string) string {
	return filepath.Join(v.OcelotDir(), "vagrant", hash)
}


func runCommandLogToChan(command *exec.Cmd, logout chan []byte, stage *build.StageUtil) error{
	stdout, _ := command.StdoutPipe()
	stderr, _ := command.StderrPipe()
	command.Start()
	//https://stackoverflow.com/questions/45922528/how-to-force-golang-to-close-opened-pipe
	go streamFromPipe(logout, stdout, stage)
	go streamFromPipe(logout, stderr, stage)
	err := command.Wait()
	return err
}


func (v *Vagrant) execute(ctx context.Context, stage *build.StageUtil, env []string, cmds []string, logout chan []byte) *pb.Result {
	cmd := exec.CommandContext(ctx, cmds[0], cmds[1:]...)
	cmd.Env = append(v.globalEnvs, env...)
	err := runCommandLogToChan(cmd, logout, stage)
	// with os/exec, if the cmd returns non-zero it returns an error so we don't have to do any explicit checking
	if err != nil {
		errMsg := fmt.Sprintf("failed to complete %s stage %s", stage.Stage, models.FAILED)
		return &pb.Result{Stage: stage.Stage, Status: pb.StageResultVal_FAIL, Error: err.Error(), Messages:[]string{errMsg}}
	}
	success := []string{fmt.Sprintf("completed %s stage %s", stage.Stage, models.CHECKMARK)}
	return &pb.Result{Stage: stage.Stage, Status: pb.StageResultVal_PASS, Error: "", Messages:success}
}



func (v *Vagrant) SetGlobalEnv(envs []string) {
	v.globalEnvs = envs
}

func (v *Vagrant) Setup(ctx context.Context, logout chan []byte, dockerId chan string, werk *pb.WerkerTask, rc cred.CVRemoteConfig, werkerPort string) (res *pb.Result, uuid string) {
	var setupMessages []string
	stage := build.InitStageUtil("SETUP")
	setupMessages = append(setupMessages, "attempting to create VM with vagrant... ")
	v.createVagrantDirec(werk.CheckoutHash)
	err := VagrantUp(ctx, v.getVagrantDirec(werk.CheckoutHash), logout, stage)
	if err != nil {
		return &pb.Result{
			Stage: stage.GetStage(),
			Status: pb.StageResultVal_FAIL,
			Error: err.Error(),
			Messages: append(setupMessages, "vagrant up command failed " + models.FAILED),
		}, v.getVagrantDirec(werk.CheckoutHash)
	}
	return &pb.Result{
		Stage: stage.GetStage(),
		Status: pb.StageResultVal_PASS,
		Error: "",
		Messages: append(setupMessages, "succesfully created vm with vagrant up " + models.CHECKMARK),
	}, v.getVagrantDirec(werk.CheckoutHash)
}

func (v *Vagrant) Execute(ctx context.Context, actions *pb.Stage, logout chan []byte, commitHash string) *pb.Result {
	return nil
}

func (v *Vagrant) ExecuteIntegration(ctx context.Context, stage *pb.Stage, stgUtil *build.StageUtil, logout chan[]byte) *pb.Result {
	return nil
}
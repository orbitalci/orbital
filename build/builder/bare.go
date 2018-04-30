package builder

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/shankj3/go-til/log"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/build/basher"
	"github.com/shankj3/ocelot/build/valet"
	cred "github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
)

type Host struct {
	*basher.Basher
	globalEnvs []string
	// host needs buildValet because it's gonna need to update the proc file at every damn turn // not sure if this is actually true now that killavalet is a thing?
	killer *valet.ContextValet
}
/*
what do we need out of bare metal builds? this is pretty much *exclusively* for any ios builds, because you can't run those in a docker container or on kubernetes.

*/

func NewHostBuilder(b *basher.Basher) build.Builder {
	return &Host{Basher: b}
}
//
//func runCommandLogToChan(command *exec.Cmd, logout chan []byte, stage *build.StageUtil) error{
//	stdout, _ := command.StdoutPipe()
//	stderr, _ := command.StderrPipe()
//	command.Start()
//	//https://stackoverflow.com/questions/45922528/how-to-force-golang-to-close-opened-pipe
//	go streamFromPipe(logout, stdout, stage)
//	go streamFromPipe(logout, stderr, stage)
//	err := command.Wait()
//	return err
//}

func (h *Host) SetGlobalEnv(envs []string) {
	h.globalEnvs = envs
}

func (h *Host) Setup(ctx context.Context, logout chan []byte, dockerIdChan chan string, werk *pb.WerkerTask, rc cred.CVRemoteConfig, werkerPort string) (*pb.Result, string) {
	close(dockerIdChan)
	var setupMessages []string
	su := build.InitStageUtil("setup")
	setupMessages = append(setupMessages, "attempting to download codebase...")
	downloadCodebase := h.execute(ctx, su, []string{}, h.DownloadCodebase(werk), logout)
	if len(downloadCodebase.Error) > 0 {
		log.Log().Error("an err happened trying to download codebase", downloadCodebase.Error)
		downloadCodebase.Messages = append(setupMessages, downloadCodebase.Messages...)
		return downloadCodebase, ""
	}
	return &pb.Result{Stage: "SETUP", Status: pb.StageResultVal_PASS, Error:"", Messages:[]string{"Running dockerless"}}, ""
}

func (h *Host) Execute(ctx context.Context, actions *pb.Stage, logout chan []byte, commitHash string) *pb.Result {
	//cmd := exec.CommandContext(ctx, )
	su := build.InitStageUtil(actions.Name)
	return h.execute(ctx, su, actions.Env, h.CDAndRunCmds(actions.Script, commitHash), logout)
}

func (h *Host) ExecuteIntegration(ctx context.Context, stage *pb.Stage, stgUtil *build.StageUtil, logout chan[]byte) *pb.Result {
	return h.execute(ctx, stgUtil, stage.Env, stage.Script, logout)
}

func (h *Host) GetContainerId() string {
	return ""
}

func (h *Host) execute(ctx context.Context, stage *build.StageUtil, env []string, cmds []string, logout chan []byte) *pb.Result {
	cmd := exec.CommandContext(ctx, cmds[0], cmds[1:]...)
	cmd.Env = append(h.globalEnvs, env...)
	err := runCommandLogToChan(cmd, logout, stage)
	// with os/exec, if the cmd returns non-zero it returns an error so we don't have to do any explicit checking
	if err != nil {
		errMsg := fmt.Sprintf("failed to complete %s stage %s", stage.Stage, models.FAILED)
		return &pb.Result{Stage: stage.Stage, Status: pb.StageResultVal_FAIL, Error: err.Error(), Messages:[]string{errMsg}}
	}
	success := []string{fmt.Sprintf("completed %s stage %s", stage.Stage, models.CHECKMARK)}
	return &pb.Result{Stage: stage.Stage, Status: pb.StageResultVal_PASS, Error: "", Messages:success}
}

func streamFromPipe(logout chan []byte, pipe io.ReadCloser, stage *build.StageUtil) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		logout <- append([]byte(stage.StageLabel), scanner.Bytes()...)
	}
	logout <- append([]byte(stage.StageLabel), []byte("Finished with stage.")...)
	log.Log().Debugf("finished writing to channel for stage %s", stage.Stage)
	if err := scanner.Err(); err != nil {
		log.IncludeErrField(err).Error("error outputting to info channel!")
		logout <- []byte("OCELOT | BY THE WAY SOMETHING WENT WRONG SCANNING STAGE INPUT FOR " + stage.Stage)
	}


}
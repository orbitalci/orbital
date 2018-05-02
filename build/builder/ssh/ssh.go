package ssh

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/shankj3/go-til/log"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/build/basher"
	"github.com/shankj3/ocelot/build/valet"
	cred "github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/common/helpers/sshhelper"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
)

type SSH struct {
	*basher.Basher
	killer 	   *valet.ContextValet
	connection *sshhelper.ContextConnection
	stage 	   *build.StageUtil
	envs	   []string
	*models.WerkerFacts
}


func NewSSHBuilder(b *basher.Basher, facts *models.WerkerFacts) build.Builder {
	return &SSH{
		Basher: b,
		connection: sshhelper.InitContextConnect(facts.Ssh.KeyFP, facts.Ssh.Password, facts.Ssh.User, facts.Ssh.Host, facts.Ssh.Port),
		WerkerFacts: facts,
	}
}

func (h *SSH) SetGlobalEnv(envs []string) {
	h.connection.SetGlobals(envs)
}

func (h *SSH) establishConnection(ctx context.Context, logout chan[]byte, setupMessages []string, stage string) *pb.Result {
	err := h.connection.Connect(ctx)
	if err != nil {
		logout <- []byte("unable to establish ssh connection")
		setupMessages = append(setupMessages, "unable to establish ssh connection " + models.FAILED)
		return &pb.Result{Stage: stage, Status: pb.StageResultVal_FAIL, Messages: setupMessages, Error: err.Error()}
	}
	if err = h.connection.CheckConnection(); err != nil {
		logout <- []byte("unable to establish ssh connection")
		setupMessages = append(setupMessages, "unable to establish ssh connection " + models.FAILED)
		return &pb.Result{Stage: stage, Status: pb.StageResultVal_FAIL, Messages: setupMessages, Error: err.Error()}
	}
	return nil
}

// Setup for the SSH werker type will send off the checkout hash as the "docker id" on the docker id channel
func (h *SSH) Setup(ctx context.Context, logout chan []byte, dockerIdChan chan string, werk *pb.WerkerTask, rc cred.CVRemoteConfig, werkerPort string) (*pb.Result, string) {
	dockerIdChan <- werk.CheckoutHash
	var setupMessages []string
	su := build.InitStageUtil("setup")
	setupMessages = append(setupMessages, "attempting to establish ssh connection...")
	if res := h.establishConnection(ctx, logout, setupMessages, su.GetStage()); res != nil {
		return res, werk.CheckoutHash
	}
	setupMessages = append(setupMessages, "successfully established ssh connection " + models.CHECKMARK)
	cmd := h.SleeplessDownloadTemplateFiles(h.RegisterIP, h.ServicePort)
	downloadTemplates := h.execute(ctx, su, []string{}, []string{cmd}, logout)
	if downloadTemplates.Status == pb.StageResultVal_FAIL {
		log.Log().Error("An error occured while trying to download templates ", downloadTemplates.Error)
		setupMessages = append(setupMessages, "failed to download templates " + models.FAILED)
		downloadTemplates.Messages = append(setupMessages, downloadTemplates.Messages...)
		return downloadTemplates, ""
	}
	setupMessages = append(setupMessages, "Set up via SSH " + models.CHECKMARK)
	return &pb.Result{Stage: su.GetStage(), Status: pb.StageResultVal_PASS, Error:"", Messages:setupMessages}, ""
}

func (h *SSH) Execute(ctx context.Context, actions *pb.Stage, logout chan []byte, commitHash string) *pb.Result {
	//cmd := exec.CommandContext(ctx, )
	su := build.InitStageUtil(actions.Name)
	return h.execute(ctx, su, actions.Env, h.CDAndRunCmds(actions.Script, commitHash), logout)
}

func (h *SSH) ExecuteIntegration(ctx context.Context, stage *pb.Stage, stgUtil *build.StageUtil, logout chan[]byte) *pb.Result {
	return h.execute(ctx, stgUtil, stage.Env, stage.Script, logout)
}

func (h *SSH) GetContainerId() string {
	return ""
}

func (h *SSH) writeToInfo(reader io.Reader, infoChan chan []byte, done chan int) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		infoChan <- append([]byte(h.stage.GetStageLabel()), scanner.Bytes()...)
	}
	//infoChan <- append([]byte(h.stage.StageLabel), []byte("Finished with stage.")...)
	log.Log().Debugf("finished writing to channel for stage %s", h.stage.Stage)
	if err := scanner.Err(); err != nil {
		log.IncludeErrField(err).Error("error outputting to info channel!")
		infoChan <- []byte("OCELOT | BY THE WAY SOMETHING WENT WRONG SCANNING STAGE INPUT FOR " + h.stage.Stage)
	}
	close(done)
}

//unwrapCommand will strip out the first two elems of the lists of cmds run in basher. ssh has a different interface than docker, and
// this is required. if it turns out that the /bin/sh -c is required _only_ for docker, we can change what the basher functions returns
func unwrapCommand(cmds []string) []string {
	return []string{cmds[2]}
}


func (h *SSH) execute(ctx context.Context, stage *build.StageUtil, env []string, cmds []string, logout chan []byte) *pb.Result {
	if len(cmds) >= 2 && strings.Contains(cmds[1], "-c") {
		cmds = unwrapCommand(cmds)
	}
	sshcmd := strings.Join(cmds, " ")
	h.stage = stage
	//defer func(){h.stage = nil}()
	err := h.connection.RunAndLog(sshcmd, env, logout, h.writeToInfo)
	if err != nil {
		errMsg := fmt.Sprintf("failed to complete %s stage %s", stage.Stage, models.FAILED)
		return &pb.Result{Stage: stage.Stage, Status: pb.StageResultVal_FAIL, Error: err.Error(), Messages:[]string{errMsg}}
	}
	success := []string{fmt.Sprintf("completed %s stage %s", stage.Stage, models.CHECKMARK)}
	return &pb.Result{Stage: stage.Stage, Status: pb.StageResultVal_PASS, Error: "", Messages:success}
}

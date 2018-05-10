package exec

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/shankj3/go-til/log"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/build/basher"
	"github.com/shankj3/ocelot/build/valet"
	cred "github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/common/helpers/exechelper"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
)

type Exec struct {
	killer     *valet.ContextValet
	stage  	   *build.StageUtil
	globalEnvs []string

	*basher.Basher
	*models.WerkerFacts
}

func NewExecBuilder(b *basher.Basher, facts *models.WerkerFacts) build.Builder {
	return &Exec{Basher:b, WerkerFacts: facts}
}

func (e *Exec) SetGlobalEnv(envs []string) {
	e.globalEnvs = envs
}


func (e *Exec) Init(ctx context.Context, hash string, logout chan[]byte) *pb.Result {
	res := &pb.Result{
		Stage: "INIT",
		Status: pb.StageResultVal_PASS,
		Messages: []string{"Initializing Exec builder... " + models.CHECKMARK},
	}
	return res
}

// Setup for the Exec werker type will send off the checkout hash as the "docker id" on the docker id channel
func (e *Exec) Setup(ctx context.Context, logout chan []byte, dockerIdChan chan string, werk *pb.WerkerTask, rc cred.CVRemoteConfig, werkerPort string) (*pb.Result, string) {
	log.Log().Infof("setting up hash %s", werk.CheckoutHash)
	dockerIdChan <- werk.CheckoutHash
	var setupMessages []string
	su := build.InitStageUtil("setup")
	cmd := e.SleeplessDownloadTemplateFiles(e.RegisterIP, e.ServicePort)
	downloadTemplates := e.execute(ctx, su, []string{}, []string{cmd}, logout)
	if downloadTemplates.Status == pb.StageResultVal_FAIL {
		log.Log().Error("An error occured while trying to download templates ", downloadTemplates.Error)
		setupMessages = append(setupMessages, "failed to download templates " + models.FAILED)
		downloadTemplates.Messages = setupMessages
		return downloadTemplates, werk.CheckoutHash
	}
	setupMessages = append(setupMessages, "Set up via Exec " + models.CHECKMARK)
	return &pb.Result{Stage: su.GetStage(), Status: pb.StageResultVal_PASS, Error:"", Messages:setupMessages}, werk.CheckoutHash
}

func (e *Exec) Execute(ctx context.Context, actions *pb.Stage, logout chan []byte, commitHash string) *pb.Result {
	//cmd := exec.CommandContext(ctx, )
	su := build.InitStageUtil(actions.Name)
	return e.execute(ctx, su, actions.Env, e.CDAndRunCmds(actions.Script, commitHash), logout)
}

func (e *Exec) ExecuteIntegration(ctx context.Context, stage *pb.Stage, stgUtil *build.StageUtil, logout chan[]byte) *pb.Result {
	return e.execute(ctx, stgUtil, stage.Env, stage.Script, logout)
}

func (e *Exec) GetContainerId() string {
	return ""
}

func (e *Exec) writeToInfo(reader io.ReadCloser, infoChan chan []byte, wg *sync.WaitGroup, readerTypeDesc string) {
	defer wg.Done()
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		infoChan <- append([]byte(e.stage.GetStageLabel()), scanner.Bytes()...)
	}
	log.Log().Debugf("finished writing to channel of %s for stage %s", readerTypeDesc, e.stage.Stage)
	if err := scanner.Err(); err != nil {
		if err != os.ErrClosed {
			log.IncludeErrField(err).Error("error outputting to info channel!")
			infoChan <- []byte(fmt.Sprintf("OCELOT | BY THE WAY SOMETHING WENT WRONG SCANNING %s STAGE INPUT FOR %s", readerTypeDesc, e.stage.Stage))
		}
	}
}

// prepCmds will reorganize the list of cmd's into a /bin/bash -c "cmds" format. it will join all the actual commands with '&&'
func prepCmds(cmds []string) (prepped [3]string) {
	if len(cmds) >= 2 && strings.Contains(cmds[1], "-c") {
		final := strings.Join(cmds[2:], " && ")
		prepped[0] = cmds[0]
		prepped[1] = cmds[1]
		prepped[2] = final
	} else {
		prepped[0] = "/bin/bash"
		prepped[1] = "-c"
		prepped[2] = strings.Join(cmds, " && ")
	}
	return
}

// execute will attempt to organize the cmd list in a /bin/sh -c format, then will execute the command while piping the stdout and stderr to the logout channel
func (e *Exec) execute(ctx context.Context, stage *build.StageUtil, env []string, cmds []string, logout chan []byte) *pb.Result {
	e.stage = stage
	var err error
	messages := []string{"starting stage " + stage.GetStage()}
	preppedCmds := prepCmds(cmds)
	command := exec.CommandContext(ctx, preppedCmds[0], preppedCmds[1:]...)
	setEnvs := append(env, e.globalEnvs...)
	// todo: should we append to the environment? with exec package, if you set the environment it overwrites _all_ other ones. do we want this behavior? or do we want to keep stuff like $HOME
	fullEnv := append(os.Environ(), setEnvs...)
	command.Env = fullEnv
	log.Log().Debugf("full env is %s", strings.Join(fullEnv, " "))
	err = exechelper.RunAndStreamCmd(command, logout, e.writeToInfo)
	if err != nil {
		return &pb.Result{
			Stage: stage.Stage,
			Status: pb.StageResultVal_FAIL,
			Error: err.Error(),
			Messages: messages,
		}
	}
	return &pb.Result{Stage: stage.Stage, Status: pb.StageResultVal_PASS, Error: "", Messages:append(messages, fmt.Sprintf("completed %s stage %s", stage.Stage, models.CHECKMARK))}
}

func (e *Exec) Close() error {
	// do nothing, there are no connections to close after the build for an exec builder
	return nil
}
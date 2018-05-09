package ssh

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

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
	killer *valet.ContextValet
	cnxn   *sshhelper.Channel
	stage  *build.StageUtil
	envs   []string
	*models.WerkerFacts
}


// NewSSHBuilder will establish the SSH connection then return the SSH builder. It will fail if it cannot
// establish a connection, as it should. It requires more than the docker builder say, because this ssh conneciton
// isn't a "clean" builder, unfortunately. It is not destroyed afterword, so we need things like hash to know what to clean up
// once the build process has completed.
func NewSSHBuilder(b *basher.Basher, facts *models.WerkerFacts) (build.Builder, error) {
	return &SSH{Basher:b, WerkerFacts: facts}, nil
}

func (h *SSH) SetGlobalEnv(envs []string) {
	h.cnxn.SetGlobals(envs)
}


func (h *SSH) Init(ctx context.Context, hash string, logout chan[]byte) *pb.Result {
	res := &pb.Result{
		Stage: "Init",
		Status: pb.StageResultVal_PASS,
	}
	cnxn, err := sshhelper.CreateSSHChannel(ctx, h.Ssh, hash)
	h.cnxn = cnxn
	if err != nil {
		res.Status = pb.StageResultVal_FAIL
		res.Error = err.Error()
		res.Messages = []string{"Failed to initialize ssh builder " + models.FAILED}
	} else {
		res.Messages = []string{"Successfully established ssh connection " + models.CHECKMARK}
	}
	return res
}

// Setup for the SSH werker type will send off the checkout hash as the "docker id" on the docker id channel
func (h *SSH) Setup(ctx context.Context, logout chan []byte, dockerIdChan chan string, werk *pb.WerkerTask, rc cred.CVRemoteConfig, werkerPort string) (*pb.Result, string) {
	log.Log().Infof("setting up hash %s", werk.CheckoutHash)
	dockerIdChan <- werk.CheckoutHash
	var setupMessages []string
	su := build.InitStageUtil("setup")
	cmd := h.SleeplessDownloadTemplateFiles(h.RegisterIP, h.ServicePort)
	downloadTemplates := h.execute(ctx, su, []string{}, []string{cmd}, logout)
	if downloadTemplates.Status == pb.StageResultVal_FAIL {
		log.Log().Error("An error occured while trying to download templates ", downloadTemplates.Error)
		setupMessages = append(setupMessages, "failed to download templates " + models.FAILED)
		downloadTemplates.Messages = setupMessages
		return downloadTemplates, werk.CheckoutHash
	}
	setupMessages = append(setupMessages, "successfully downloaded templates " + models.CHECKMARK)
	return &pb.Result{Stage: su.GetStage(), Status: pb.StageResultVal_PASS, Error:"", Messages:setupMessages}, werk.CheckoutHash
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

func (h *SSH) writeToInfo(reader io.Reader, infoChan chan []byte, wg *sync.WaitGroup) {
	defer wg.Done()
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
	err := h.cnxn.RunAndLog(sshcmd, env, logout, h.writeToInfo)
	if err != nil {
		errMsg := fmt.Sprintf("failed to complete %s stage %s", stage.Stage, models.FAILED)
		return &pb.Result{Stage: stage.Stage, Status: pb.StageResultVal_FAIL, Error: err.Error(), Messages:[]string{errMsg}}
	}
	success := []string{fmt.Sprintf("completed %s stage %s", stage.Stage, models.CHECKMARK)}
	return &pb.Result{Stage: stage.Stage, Status: pb.StageResultVal_PASS, Error: "", Messages:success}
}


func (h *SSH) Close() error {
	if h.cnxn != nil {
		return h.cnxn.Close()
	}
	return nil
}
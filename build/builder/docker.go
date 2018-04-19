package builder

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	ocelog "bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/go-til/vault"

	"bitbucket.org/level11consulting/ocelot/build"
	"bitbucket.org/level11consulting/ocelot/build/basher"
	cred "bitbucket.org/level11consulting/ocelot/common/credentials"
	"bitbucket.org/level11consulting/ocelot/common/helpers/dockrhelper"
	"bitbucket.org/level11consulting/ocelot/models/pb"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type Docker struct{
	Log	io.ReadCloser
	ContainerId	string
	DockerClient *client.Client
	*basher.Basher
}

func NewDockerBuilder(b *basher.Basher) build.Builder {
	return &Docker{nil, "", nil, b}
}

func (d *Docker) GetContainerId() string {
	return d.ContainerId
}

func (d *Docker) Setup(ctx context.Context, logout chan []byte, dockerIdChan chan string, werk *pb.WerkerTask, rc cred.CVRemoteConfig, werkerPort string) (*pb.Result, string) {
	var setupMessages []string

	su := build.InitStageUtil("setup")

	logout <- []byte(su.GetStageLabel() + "Setting up...")
	cli, err := client.NewEnvClient()
	d.DockerClient = cli

	if err != nil {
		ocelog.Log().Debug("returning failed stage because could not create docker env client")
		return &pb.Result {
			Stage:  su.GetStage(),
			Status: pb.StageResultVal_FAIL,
			Error: err.Error(),
		}, ""
	}
	imageName := werk.BuildConf.Image

	out, err := dockrhelper.RobustImagePull(imageName)
	if err != nil {
		return &pb.Result{
			Stage:  su.GetStage(),
			Status: pb.StageResultVal_FAIL,
			Error:  err.Error(),
			Messages: setupMessages,
		}, ""
	}
	defer out.Close()
	setupMessages = append(setupMessages, fmt.Sprintf("pulled image %s \u2713", imageName))
	bufReader := bufio.NewReader(out)
	d.writeToInfo(su.GetStageLabel(), bufReader, logout)

	logout <- []byte(su.GetStageLabel() + "Creating container...")

	//add environment variables that will always be avilable on the machine - GIT_HASH, BUILD_ID, GIT_HASH_SHORT, GIT_BRANCH
	paddedEnvs := []string{fmt.Sprintf("GIT_HASH=%s", werk.CheckoutHash), fmt.Sprintf("BUILD_ID=%d", werk.Id), fmt.Sprintf("GIT_HASH_SHORT=%s", werk.CheckoutHash[:7]), fmt.Sprintf("GIT_BRANCH=%s", werk.Branch)}
	paddedEnvs = append(paddedEnvs, werk.BuildConf.Env...)


	//container configurations
	containerConfig := &container.Config{
		Image: imageName,
		User: "root",
		Env: paddedEnvs,
		Cmd: d.DownloadTemplateFiles(werkerPort),
		AttachStderr: true,
		AttachStdout: true,
		AttachStdin:true,
		Tty:true,
	}

	//homeDirectory, _ := homedir.Expand("~/.ocelot")
	//host config binds are mount points
	hostConfig := &container.HostConfig{
		//TODO: have it be overridable via env variable
		Binds: []string{"/var/run/docker.sock:/var/run/docker.sock"},
		//Binds: []string{ homeDirectory + ":/.ocelot", "/var/run/docker.sock:/var/run/docker.sock"},
		NetworkMode: "host",
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig , hostConfig, nil, "")


	if err != nil {
		ocelog.IncludeErrField(err).Error("returning failed because could not create container")
		return &pb.Result{
			Stage:  su.GetStage(),
			Status: pb.StageResultVal_FAIL,
			Error:  err.Error(),
			Messages: setupMessages,
		}, ""
	}

	setupMessages = append(setupMessages, fmt.Sprint("created build container \u2713"))

	for _, warning := range resp.Warnings {
		logout <- []byte(warning)
	}

	//TODO: is creating a channel to use it once overkill....
	ocelog.Log().Debug("sweet, sending container id back and closing dockeruuid chan")
	dockerIdChan <- resp.ID
	close(dockerIdChan)

	logout <- []byte(su.GetStageLabel() + "Container created with ID " + resp.ID)

	d.ContainerId = resp.ID
	ocelog.Log().Debug("starting up container")
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		ocelog.IncludeErrField(err).Error("returning failed because could not start container")
		return &pb.Result{
			Stage:  su.GetStage(),
			Status: pb.StageResultVal_FAIL,
			Error:  err.Error(),
			Messages: setupMessages,
		}, ""
	}

	logout <- []byte(su.GetStageLabel()  + "Container " + resp.ID + " started")


	//since container is created in setup, log tailing via container is also kicked off in setup
	containerLog, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow: true,
	})

	if err != nil {
		ocelog.IncludeErrField(err).Error("returning failed setup because could not get logs of container")
		return &pb.Result{
			Stage: su.GetStage(),
			Status: pb.StageResultVal_FAIL,
			Error:  err.Error(),
			Messages: setupMessages,
		}, d.ContainerId
	}

	d.Log = containerLog
	bufReader = bufio.NewReader(containerLog)

	d.writeToInfo(su.GetStageLabel() , bufReader, logout)
	setupMessages = append(setupMessages, "attempting to download ocelot package dependencies...")
	installed := d.Exec(ctx, su.GetStage(), su.GetStageLabel(), []string{}, d.InstallPackageDeps(), logout)
	if len(installed.Error) > 0 {
		ocelog.Log().Error("an error happened installing package deps ", installed.Error)
		installed.Messages = append(setupMessages, installed.Messages...)
		return installed, d.ContainerId
	}
	setupMessages = append(setupMessages, "attempting to download codebase...")
	downloadCodebase := d.Exec(ctx, su.GetStage(), su.GetStageLabel(), []string{}, d.DownloadCodebase(werk), logout)
	if len(downloadCodebase.Error) > 0 {
		ocelog.Log().Error("an err happened trying to download codebase", downloadCodebase.Error)
		downloadCodebase.Messages = append(setupMessages, downloadCodebase.Messages...)
		return downloadCodebase, d.ContainerId
	}


	logout <- []byte(su.GetStageLabel()  + "Retrieving SSH Key")


	acctName := strings.Split(werk.FullName, "/")[0]
	vaultAddr := d.getVaultAddr(rc.GetVault())
	ocelog.Log().Info("ADDRESS FOR VAULT IS: " + vaultAddr)

	setupMessages = append(setupMessages, fmt.Sprintf("downloading SSH key for %s...", werk.FullName))
	sctType := pb.SubCredType(werk.VcsType)
	identifier, _ := pb.CreateVCSIdentifier(sctType, acctName)
	ocelog.Log().Debug("identifier is ", identifier)
	result := d.Exec(ctx, su.GetStage(), su.GetStageLabel(), []string{"VAULT_ADDR="+vaultAddr}, d.DownloadSSHKey(
		werk.VaultToken,
		cred.BuildCredPath(sctType, acctName, pb.CredType_VCS, identifier)), logout)
	if len(result.Error) > 0 {
		ocelog.Log().Error("an err happened trying to download ssh key", result.Error)
		result.Messages = append(setupMessages, result.Messages...)
		return result, d.ContainerId
	}

	setupMessages = append(setupMessages, fmt.Sprintf("successfully downloaded SSH key for %s  \u2713", werk.FullName), "completed setup stage \u2713")
	result.Messages = setupMessages
	return result, d.ContainerId
}

func (d *Docker) getVaultAddr(vaulty vault.Vaulty) string {
	registerdAddr := vaulty.GetAddress()
	// if on localhost, set to basher's loopback ip
	if strings.Contains(registerdAddr, "127.0.0.1") {
		start := strings.Index(registerdAddr, "127.0.0.1")
		end := start + 9
		registerdAddr = registerdAddr[:start] + d.Basher.LoopbackIp + registerdAddr[end:]
		ocelog.Log().Info("using build loopback address ", registerdAddr)
	}
	return registerdAddr
}


// ExecuteIntegration will basically run Execute but without the cd and run cmds because we are generating the scripts in the code
func (d *Docker) ExecuteIntegration(ctx context.Context, stage *pb.Stage, stgUtil *build.StageUtil, logout chan[]byte) *pb.Result {
	return d.Exec(ctx, stgUtil.GetStage(), stgUtil.GetStageLabel(), stage.Env, stage.Script, logout)
}


func (d *Docker) Execute(ctx context.Context, stage *pb.Stage, logout chan []byte, commitHash string) *pb.Result {
	if len(d.ContainerId) == 0 {
		return &pb.Result {
			Stage: stage.Name,
			Status: pb.StageResultVal_FAIL,
			Error: "no container exists, setup before executing",
		}
	}

	su := build.InitStageUtil(stage.Name)
	return d.Exec(ctx, su.GetStage(), su.GetStageLabel(), stage.Env, d.CDAndRunCmds(stage.Script, commitHash), logout)
}

func (d *Docker) Exec(ctx context.Context, currStage string, currStageStr string, env []string, cmds []string, logout chan []byte) *pb.Result {
	var stageMessages []string
	resp, err := d.DockerClient.ContainerExecCreate(ctx, d.ContainerId, types.ExecConfig{
		Tty: true,
		AttachStdin: true,
		AttachStderr: true,
		AttachStdout: true,
		Env: env,
		Cmd: cmds,
	})
	if err != nil {
		return &pb.Result{
			Stage:  currStage,
			Status: pb.StageResultVal_FAIL,
			Error:  err.Error(),
			Messages: stageMessages,
		}
	}

	attachedExec, err := d.DockerClient.ContainerExecAttach(ctx, resp.ID, types.ExecConfig{
		Tty: true,
		AttachStdin: true,
		AttachStderr: true,
		AttachStdout: true,
		Env: env,
		Cmd: cmds,
	})

	defer attachedExec.Conn.Close()

	d.writeToInfo(currStageStr, attachedExec.Reader, logout)
	inspector, err := d.DockerClient.ContainerExecInspect(ctx, resp.ID)

	// todo: have stage have exit code in case a stage doesn't care if exit code is nonzero (tj recommendation)
	if inspector.ExitCode != 0 || err != nil {
		stageMessages = append(stageMessages, fmt.Sprintf("failed to complete %s stage \u2717", currStage))
		var errStr string
		if err == nil {
			errStr = "exit code was not 0"
		} else {
			errStr = err.Error()
		}

		return &pb.Result{
			Stage: currStage,
			Status: pb.StageResultVal_FAIL,
			Error: errStr,
			Messages: stageMessages,
		}
	}
	stageMessages = append(stageMessages, fmt.Sprintf("completed %s stage \u2713", currStage))
	return &pb.Result{
		Stage:  currStage,
		Status: pb.StageResultVal_PASS,
		Error:  "",
		Messages: stageMessages,
	}
}

func (d *Docker) writeToInfo(stage string, rd *bufio.Reader, infochan chan []byte) {
	scanner := bufio.NewScanner(rd)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		str := string(scanner.Bytes())
		infochan <- []byte(stage + str)
		//our setup script will echo this to stdout, telling us script is finished downloading. This is HACK for keeping container alive
		if strings.Contains(str, "Ocelot has finished with downloading templates") {
			ocelog.Log().Debug("finished with source code, returning out of writeToInfo")
			return
		}
	}
	ocelog.Log().Debug("finished writing to channel for stage ", stage)
	if err := scanner.Err(); err != nil {
		ocelog.IncludeErrField(err).Error("error outputing to info channel!")
		infochan <- []byte("OCELOT | BY THE WAY SOMETHING WENT WRONG SCANNING STAGE INPUT")
	}
}
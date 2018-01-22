package builder

import (
	ocelog "bitbucket.org/level11consulting/go-til/log"
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/mitchellh/go-homedir"
	"io"
	"strings"

	//"os/exec"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"strings"
)

type Docker struct{
	Log	io.ReadCloser
	ContainerId	string
	DockerClient *client.Client
	RemoteConfig cred.CVRemoteConfig
	*Basher
}

func NewDockerBuilder(b *Basher) Builder {
	remoteCred, _ := cred.New()
	return &Docker{nil, "", nil, remoteCred, b}
}

func (d *Docker) Setup(logout chan []byte, werk *pb.WerkerTask) *Result {
	su := InitStageUtil("setup")

	logout <- []byte(su.GetStageLabel() + "Setting up...")

	ctx := context.Background()
	cli, err := client.NewEnvClient()
	d.DockerClient = cli

	if err != nil {
		return &Result{
			Stage:  su.GetStage(),
			Status: FAIL,
			Error:  err,
		}
	}

	imageName := werk.BuildConf.Image

	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't pull image!")
		return &Result{
			Stage:  su.GetStage(),
			Status: FAIL,
			Error:  err,
		}
	}
	var byt []byte
	buf := bufio.NewReader(out)
	buf.Read(byt)
	fmt.Println(string(byt))
	defer out.Close()

	bufReader := bufio.NewReader(out)
	d.writeToInfo(su.GetStageLabel(), bufReader, logout)

	logout <- []byte(su.GetStageLabel() + "Creating container...")


	//container configurations
	containerConfig := &container.Config{

		Image: imageName,
		Env: werk.BuildConf.Env,
		Cmd: d.DownloadCodebase(werk),
		AttachStderr: true,
		AttachStdout: true,
		AttachStdin:true,
		Tty:true,
	}

	homeDirectory, _ := homedir.Expand("~/.ocelot")
	//host config binds are mount points
	hostConfig := &container.HostConfig{
		//TODO: have it be overridable via env variable
		Binds: []string{ homeDirectory + ":/.ocelot", "/var/run/docker.sock:/var/run/docker.sock"},
		NetworkMode: "host",
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig , hostConfig, nil, "")

	if err != nil {
		return &Result{
			Stage:  su.GetStage(),
			Status: FAIL,
			Error:  err,
		}
	}

	for _, warning := range resp.Warnings {
		logout <- []byte(warning)
	}

	logout <- []byte(su.GetStageLabel() + "Container created with ID " + resp.ID)

	d.ContainerId = resp.ID

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return &Result{
			Stage:  su.GetStage(),
			Status: FAIL,
			Error:  err,
		}
	}

	logout <- []byte(su.GetStageLabel()  + "Container " + resp.ID + " started")

	//since container is created in setup, log tailing via container is also kicked off in setup
	containerLog, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow: true,
	})

	if err != nil {
		return &Result{
			Stage: su.GetStage(),
			Status: FAIL,
			Error:  err,
		}
	}

	d.Log = containerLog
	bufReader = bufio.NewReader(containerLog)
	d.writeToInfo(su.GetStageLabel() , bufReader, logout)

	return &Result{
		Stage:  su.GetStage(),
		Status: PASS,
		Error:  nil,
	}
}

func (d *Docker) Cleanup(logout chan []byte) {
	su := InitStageUtil("cleanup")
	logout <- []byte(su.GetStageLabel() + "Performing build cleanup...")

	//TODO: review, should we be creating new contexts for every stage?
	cleanupCtx := context.Background()
	if d.Log != nil {
		d.Log.Close()
	}
	if err := d.DockerClient.ContainerKill(cleanupCtx, d.ContainerId, "SIGKILL"); err != nil {
		ocelog.IncludeErrField(err).WithField("containerId", d.ContainerId).Error("couldn't kill")
	} else {
		if err := d.DockerClient.ContainerRemove(cleanupCtx, d.ContainerId, types.ContainerRemoveOptions{}); err != nil {
			ocelog.IncludeErrField(err).WithField("containerId", d.ContainerId).Error("couldn't rm")
		}
	}
	d.DockerClient.Close()
}


func (d *Docker) Execute(stage *pb.Stage, logout chan []byte, commitHash string) *Result {
	if len(d.ContainerId) == 0 {
		return &Result {
			Stage: stage.Name,
			Status: FAIL,
			Error: errors.New("no container exists, setup before executing"),
		}
	}

	su := InitStageUtil(stage.Name)
	return d.Exec(su.GetStage(), su.GetStageLabel(), stage.Env, d.BuildScript(stage.Script, commitHash), logout)
}

//uses the repo creds from admin to store artifact - keyed by acctname
func (d *Docker) SaveArtifact(logout chan []byte, task *pb.WerkerTask, commitHash string) *Result {
	su := &StageUtil{
		Stage: "SaveArtifact",
		StageLabel: "SAVE_ARTIFACT | ",
	}

	logout <- []byte(su.GetStageLabel() + "Saving artifact to ...")

	if len(d.ContainerId) == 0 {
		return &Result {
			Stage: su.GetStage(),
			Status: FAIL,
			Error: errors.New("no container exists, setup before executing"),
		}
	}

	//check if build tool if set to maven (cause that's the only thing that we use to push to nexus right now)
	if strings.Compare(task.BuildConf.BuildTool, "maven") != 0 {
		logout <- []byte(fmt.Sprintf(su.GetStageLabel() + "build tool %s not part of accepted values: %s...", task.BuildConf.BuildTool, "maven"))
		return &Result {
			Stage: su.GetStage(),
			Status: FAIL,
			Error: errors.New(fmt.Sprintf("build tool %s not part of accepted values: %s...", task.BuildConf.BuildTool, "maven")),
		}
	}

	//check if nexus creds exist
	err := d.RemoteConfig.CheckExists(cred.BuildCredPath("nexus", task.AcctName, cred.Repo))

	if err != nil {
		logout <- []byte(su.GetStageLabel() + err.Error())
		return &Result {
			Stage: su.GetStage(),
			Status: FAIL,
			Error: errors.New("nexus credentials don't exist for " + task.AcctName),
		}
	}

	return d.Exec(su.GetStage(), su.GetStageLabel(), nil, d.PushToNexus(commitHash), logout)
}


func (d *Docker) Exec(currStage string, currStageStr string, env []string, cmds []string, logout chan []byte) *Result {
	ctx := context.Background()

	resp, err := d.DockerClient.ContainerExecCreate(ctx, d.ContainerId, types.ExecConfig{
		Tty: true,
		AttachStdin: true,
		AttachStderr: true,
		AttachStdout: true,
		Env: env,
		Cmd: cmds,
	})

	if err != nil {
		return &Result{
			Stage:  currStage,
			Status: FAIL,
			Error:  err,
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

	if err != nil {
		return &Result{
			Stage:  currStage,
			Status: FAIL,
			Error:  err,
		}
	}
	return &Result{
		Stage:  currStage,
		Status: PASS,
		Error:  nil,
	}
}


func (d *Docker) writeToInfo(stage string, rd *bufio.Reader, infochan chan []byte) {
	for {
		//TODO: if we swap to scanner will it outputs nicer?
		str, err := rd.ReadString('\n')

		if err != nil {
			if err != io.EOF {
				ocelog.Log().Error("Read Error:", err)
			} else {
				ocelog.Log().Debug("EOF, finished writing to Info")
			}
			return
		}

		infochan <- []byte(stage + str)

		//our setup script will echo this to stdout, telling us script is finished downloading. This is HACK for keeping container alive
		if str == "Finished with downloading source code\r\n" {
			return
		}
	}
}

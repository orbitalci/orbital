package builder

import (
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"bufio"
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"io"
	"github.com/mitchellh/go-homedir"
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"errors"
)

type Docker struct{
	Log	io.ReadCloser
	ContainerId	string
	DockerClient *client.Client
	*Basher
}

func NewDockerBuilder(b *Basher) Builder {
	return &Docker{nil, "", nil, b}
}

func (d *Docker) Setup(logout chan []byte, werk *pb.WerkerTask) *Result {
	//TODO: do some sort of util for the stage name + formatting
	stage := "setup"
	stagePrintln := "SETUP | "

	logout <- []byte(stagePrintln + "Setting up...")

	ctx := context.Background()
	cli, err := client.NewEnvClient()
	d.DockerClient = cli

	if err != nil {
		return &Result{
			Stage:  stage,
			Status: FAIL,
			Error:  err,
		}
	}

	imageName := werk.BuildConf.Image

	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't pull image!")
		return &Result{
			Stage:  stage,
			Status: FAIL,
			Error:  err,
		}
	}
	defer out.Close()

	bufReader := bufio.NewReader(out)
	d.writeToInfo(stagePrintln, bufReader, logout)

	logout <- []byte(stagePrintln + "Creating container...")


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
	//host configs like mount points
	hostConfig := &container.HostConfig{
		//TODO: have it be overridable via env variable
		Binds: []string{ homeDirectory + ":/.ocelot"},
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig , hostConfig, nil, "")

	if err != nil {
		return &Result{
			Stage:  stage,
			Status: FAIL,
			Error:  err,
		}
	}

	for _, warning := range resp.Warnings {
		logout <- []byte(warning)
	}

	logout <- []byte(stagePrintln + "Container created with ID " + resp.ID)

	d.ContainerId = resp.ID

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return &Result{
			Stage:  stage,
			Status: FAIL,
			Error:  err,
		}
	}

	logout <- []byte(stagePrintln + "Container " + resp.ID + " started")

	//since container is created in setup, log tailing via container is also kicked off in setup
	containerLog, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow: true,
	})

	if err != nil {
		return &Result{
			Stage: stage,
			Status: FAIL,
			Error:  err,
		}
	}

	d.Log = containerLog
	bufReader = bufio.NewReader(containerLog)
	d.writeToInfo(stagePrintln, bufReader, logout)

	return &Result{
		Stage:  stage,
		Status: PASS,
		Error:  nil,
	}
}

func (d *Docker) Cleanup() {
	//TODO: review, should we be creating new contexts for every stage?
	cleanupCtx := context.Background()

	d.Log.Close()
	//TODO: review kill and remove options with jessi + tj, dunno what these things do, but must be doing something wrong if this doesn't kill the container...
	d.DockerClient.ContainerKill(cleanupCtx, d.ContainerId, "SIGTERM")
	d.DockerClient.ContainerRemove(cleanupCtx, d.ContainerId, types.ContainerRemoveOptions{})
	d.DockerClient.Close()
}

//TODO: depending on how Execute function turns out, this may end up being merged into that and we case switch in Execute
//TODO: if type is build, since build = deploy afterwards
func (d *Docker) Build(logout chan []byte, stage *pb.Stage, commitHash string) *Result {
	currStage := "build"
	currStageStr := "BUILD | "

	logout <- []byte(currStageStr + "Building...")

	if len(d.ContainerId) == 0 {
		return &Result {
			Stage: currStage,
			Status: FAIL,
			Error: errors.New("no container exists, setup before executing"),
		}
	}

	ctx := context.Background()

	resp, err := d.DockerClient.ContainerExecCreate(ctx, d.ContainerId, types.ExecConfig{
		Tty: true,
		AttachStdin: true,
		AttachStderr: true,
		AttachStdout: true,
		Env: stage.Env,
		Cmd: d.BuildAndDeploy(stage.Script, commitHash),
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
		Env: stage.Env,
		Cmd: d.BuildAndDeploy(stage.Script, commitHash),
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

//TODO: actually write the code that executes scripts from other stages
func (d *Docker) Execute(stage string, actions *pb.Stage, logout chan []byte) *Result {
	if len(d.ContainerId) == 0 {
		return &Result {
			Stage: stage,
			Status: FAIL,
			Error: errors.New("No container exists, setup before executing"),
		}
	}

	return &Result{

	}
}


func (d *Docker) writeToInfo(stage string, rd *bufio.Reader, infochan chan []byte) {
	for {
		//TODO: if we swap to scanner will it outputs nicer?
		str, err := rd.ReadString('\n')

		if err != nil {
			ocelog.Log().Info("Read Error:", err)
			return
		}

		infochan <- []byte(stage + str)

		//our setup script will echo this to stdout, telling us script is finished downloading. This is HACK for keeping container alive
		if str == "Finished with downloading source code\r\n" {
			return
		}
	}
}

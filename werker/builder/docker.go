package builder

import (
	ocelog "bitbucket.org/level11consulting/go-til/log"
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"bufio"
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/mitchellh/go-homedir"
	"io"
	"strings"
	"fmt"
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
	setupMessages := []string{}
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
			Messages: setupMessages,
		}
	}
	setupMessages = append(setupMessages, fmt.Sprintf("pulled image %s \u2713", imageName))
	//var byt []byte
	//buf := bufio.NewReader(out)
	//buf.Read(byt)
	//fmt.Println(string(byt))
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
	//host config binds are mount points
	hostConfig := &container.HostConfig{
		//TODO: have it be overridable via env variable
		Binds: []string{ homeDirectory + ":/.ocelot", "/var/run/docker.sock:/var/run/docker.sock"},
		NetworkMode: "host",
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig , hostConfig, nil, "")


	if err != nil {
		return &Result{
			Stage:  stage,
			Status: FAIL,
			Error:  err,
			Messages: setupMessages,
		}
	}

	setupMessages = append(setupMessages, fmt.Sprint("created build container \u2713"))

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
			Messages: setupMessages,
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
			Messages: setupMessages,
		}
	}

	d.Log = containerLog
	bufReader = bufio.NewReader(containerLog)
	d.writeToInfo(stagePrintln, bufReader, logout)

	setupMessages = append(setupMessages, "completed setup stage \u2713")
	return &Result{
		Stage:  stage,
		Status: PASS,
		Error:  nil,
		Messages: setupMessages,
	}
}

func (d *Docker) Cleanup() {
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
	stageMessages := []string{}

	if len(d.ContainerId) == 0 {
		return &Result {
			Stage: stage.Name,
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
			Stage:  stage.Name,
			Status: FAIL,
			Error:  err,
			Messages: stageMessages,
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

	d.writeToInfo(strings.ToUpper(stage.Name) + " | ", attachedExec.Reader, logout)
	inspector, err := d.DockerClient.ContainerExecInspect(ctx, resp.ID)


	if inspector.ExitCode != 0 || err != nil {
		stageMessages = append(stageMessages, fmt.Sprintf("failed to complete %s stage \u2717", stage.Name))
		return &Result{
			Stage: stage.Name,
			Status: FAIL,
			Error: err,
			Messages: stageMessages,
		}
	}

	stageMessages = append(stageMessages, fmt.Sprintf("completed %s stage \u2713", stage.Name))
	return &Result{
		Stage:  stage.Name,
		Status: PASS,
		Error:  nil,
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
		if strings.Contains(str, "Ocelot has finished with downloading source code") {
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

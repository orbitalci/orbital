package builder

import (
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"bufio"
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"io"
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"errors"
	"strings"
)

type Docker struct{
	Log	io.ReadCloser
	ContainerId	string
	DockerClient *client.Client
}

func NewDockerBuilder() Builder {
	return &Docker{}
}

func (d *Docker) Setup(logout chan []byte, image string, globalEnvs []string, setupCmds []string) *Result {
	currentStage := "SETUP | "

	ocelog.Log().Debug("doing the setup")
	ctx := context.Background()

	cli, err := client.NewEnvClient()
	d.DockerClient = cli

	if err != nil {
		return &Result{
			Stage:  "setup",
			Status: FAIL,
			Error:  err,
		}
	}

	imageName := image

	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	defer out.Close()

	if err != nil {
		return &Result{
			Stage:  "setup",
			Status: FAIL,
			Error:  err,
		}
	}

	bufReader := bufio.NewReader(out)
	d.writeToInfo(currentStage, bufReader, logout)

	logout <- []byte(currentStage + "Creating container...")

	//container configurations
	containerConfig := &container.Config{
		Image: imageName,
		Env: globalEnvs,
		Cmd: setupCmds,
		AttachStderr: true,
		AttachStdout: true,
		AttachStdin:true,
		Tty:true,
	}

	//host configs like mount points
	hostConfig := &container.HostConfig{
		Binds: []string{"/Users/mariannefeng/.ocelot:/.ocelot"},
		//Binds: []string{"/home/mariannefeng/.ocelot:/.ocelot"},
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig , hostConfig, nil, "")

	if err != nil {
		return &Result{
			Stage:  "setup",
			Status: FAIL,
			Error:  err,
		}
	}

	for _, warning := range resp.Warnings {
		logout <- []byte(warning)
	}
	logout <- []byte(currentStage + "Container created with ID " + resp.ID)
	d.ContainerId = resp.ID

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return &Result{
			Stage:  "setup",
			Status: FAIL,
			Error:  err,
		}
	}

	logout <- []byte(currentStage + "Container " + resp.ID + " started")

	//since container is created in setup, log tailing via container is also kicked off in setup
	containerLog, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow: true,
		})

	if err != nil {
		return &Result{
			Stage:  "setup",
			Status: FAIL,
			Error:  err,
		}
	}

	d.Log = containerLog
	bufReader = bufio.NewReader(containerLog)
	go d.writeToInfo(currentStage, bufReader, logout)

	return &Result{
		Stage:  "setup",
		Status: PASS,
		Error:  nil,
	}
}

func (d *Docker) Cleanup() {
	//d.Log.Close()
	//TODO: destroy container
	//d.DockerClient.Close()
}

func (d *Docker) Build(logout chan []byte, envs []string, cmds []string, commitHash string) *Result {
	ocelog.Log().Debug("inside of build function now")

	if len(d.ContainerId) == 0 {
		return &Result {
			Stage: "build",
			Status: FAIL,
			Error: errors.New("No container exists, setup before executing"),
		}
	}
	ctx := context.Background()

	concatCmd := append([]string{"cd", commitHash, "&&"}, cmds...)
	completeCmd := append([]string{"/bin/sh", "-c"}, "'" + strings.Join(concatCmd, " ") + "'")

	resp, err := d.DockerClient.ContainerExecCreate(ctx, d.ContainerId, types.ExecConfig{
		Tty: true,
		AttachStdin: true,
		AttachStderr: true,
		AttachStdout: true,
		Env: envs,
		Cmd: completeCmd,
	})

	if err != nil {
		return &Result{
			Stage:  "build",
			Status: FAIL,
			Error:  err,
		}
	}

	attachedExec, err := d.DockerClient.ContainerExecAttach(ctx, resp.ID, types.ExecConfig{
		Tty: true,
		AttachStdin: true,
		AttachStderr: true,
		AttachStdout: true,
		Env: envs,
		Cmd: completeCmd,
	})

	inspect, err := d.DockerClient.ContainerExecInspect(ctx, resp.ID)
	ocelog.Log().Debug(inspect)

	//TODO: close conn
	go d.writeToInfo("BUILD | ", attachedExec.Reader, logout)

	if err != nil {
		return &Result{
			Stage:  "build",
			Status: FAIL,
			Error:  err,
		}
	}

	return &Result{
		Stage:  "build",
		Status: PASS,
		Error:  nil,
	}
}

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
		str, err := rd.ReadString('\n')
		if err != nil {
			ocelog.Log().Info("Read Error:", err)
			return
		}
		infochan <- []byte(stage + str)
	}
}

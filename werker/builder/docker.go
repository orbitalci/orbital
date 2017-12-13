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
)

type Docker struct{
	Log	io.ReadCloser
	ContainerId	string
	DockerClient *client.Client
}

func NewDockerBuilder() Builder {
	return &Docker{}
}

func (d *Docker) Setup(logout chan []byte, image string, globalEnvs []string) *Result {
	currentStage := "SETUP | "

	ocelog.Log().Debug("doing setup")
	ctx := context.Background()

	cli, err := client.NewEnvClient()
	d.DockerClient = cli

	if err != nil {
		panic(err)
	}

	imageName := image

	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	defer out.Close()

	if err != nil {
		return &Result{
			Stage:  "Setup",
			Status: FAIL,
			Error:  err,
		}
	}

	bufReader := bufio.NewReader(out)
	d.writeToInfo(currentStage, bufReader, logout)

	//TODO: run bash script for kicking off cmd
	logout <- []byte(currentStage + "Creating container...")
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Env: globalEnvs,
	}, nil, nil, "")

	if err != nil {
		return &Result{
			Stage:  "Setup",
			Status: FAIL,
			Error:  err,
		}
	}

	logout <- []byte(currentStage + "Container created with ID " + resp.ID)
	d.ContainerId = resp.ID

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return &Result{
			Stage:  "Setup",
			Status: FAIL,
			Error:  err,
		}
	}

	logout <- []byte(currentStage + "Container " + resp.ID + " started")

	//since container is created in setup, log tailing via container is also kicked off in setup
	containerLog, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	d.Log = &containerLog
	bufReader = bufio.NewReader(containerLog)
	d.writeToInfo(currentStage, bufReader, logout)
	if err != nil {
		return &Result{
			Stage:  "Setup",
			Status: FAIL,
			Error:  err,
		}
	}

	return &Result{
		Stage:  "Setup",
		Status: PASS,
		Error:  nil,
	}
}

func (d *Docker) Cleanup() {
	d.Log.Close()
	//TODO: destroy container
}

func (d *Docker) Build(logout chan []byte) *Result {
	return &Result{}
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

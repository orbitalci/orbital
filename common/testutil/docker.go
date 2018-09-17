package testutil

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// DockerCreateExec creates a docker container by running the docker client via the os/exec package. It will return a cleanup function that will
// kill the container.
func DockerCreateExec(t *testing.T, ctx context.Context, imageName string, ports []string, mounts ...string) (cleanup func(), err error) {
	portsString := " -p " + strings.Join(ports, " -p ")
	mountsStrings := " -v " + strings.Join(mounts, " -v ")
	command := fmt.Sprintf("docker run --rm -d %s %s %s", portsString, mountsStrings, imageName)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Log(stderr.String())
		t.Fatal(err)
	}
	time.Sleep(2 * time.Second)
	t.Log(stdout.String())
	id := strings.TrimSpace(stdout.String())
	cleanup = func() {
		cmd := exec.Command("/bin/bash", "-c", "docker kill "+id)
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
	}
	return cleanup, nil
}

// i used all the damn functions that docker is using
// WHY
func DockerCreate(t *testing.T, ctx context.Context, imageName string, ports []string, mounts ...string) (isRunning bool, cleanup func()) {
	cli, err := client.NewEnvClient()
	if err != nil {
		t.Fatal("couldn't create docker cli, err: ", err.Error())
	}
	exposedPorts, bindings, err := nat.ParsePortSpecs(ports)
	if err != nil {
		t.Error(err)
		return false, nil
	}
	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		t.Fatal("couldn't pull image, err: ", err.Error())
	}
	defer out.Close()
	containerConfig := &container.Config{
		Image:        imageName,
		ExposedPorts: exposedPorts,
		AttachStderr: true,
		AttachStdout: true,
		AttachStdin:  true,
		Tty:          true,
	}
	binds := append([]string{"/var/run/docker.sock:/var/run/docker.sock"}, mounts...)
	hostConfig := &container.HostConfig{
		//TODO: have it be overridable via env variable
		Binds:        binds,
		PortBindings: bindings,
		AutoRemove:   true,
		//Binds: []string{ homeDirectory + ":/.ocelot", "/var/run/docker.sock:/var/run/docker.sock"},
		NetworkMode: "host",
	}
	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, "")
	if err != nil {
		t.Error("could not create container, error: ", err.Error(), "\nwarnings: ", strings.Join(resp.Warnings, "\n  - "))
		return
	}
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		t.Error("couldnt create container, error: ", err.Error())
		return
	}
	//containerLog, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
	//	ShowStdout: true,
	//	ShowStderr: true,
	//	Follow:     true,
	//})
	//if err != nil {
	//	t.Error("couldn't get container log, error: ", err.Error())
	//	return
	//}
	cleanup = func() { cli.ContainerKill(ctx, resp.ID, "SIGKILL") }
	isRunning = true
	return
}

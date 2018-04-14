package builder

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	cleaner2 "bitbucket.org/level11consulting/ocelot/newocy/build/cleaner"
)

func tarTemplates(t *testing.T) func(t *testing.T) {
	//tar -cvf werker_files.tar *
	here, _ := ioutil.ReadDir(".")
	fmt.Println(here)
	cmd := exec.Command("/bin/sh", "-c", "tar -cvf ../werker_files.tar *")
	cmd.Dir = "./template/"
	var out, err bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &err
	errr := cmd.Run()
	t.Log(out.String())
	if errr != nil {
		t.Fatal(fmt.Sprintf("unable to tar up template direc, stdout: %s \n stderr: %s \n err: %s", out.String(), err.String(), errr.Error()))
	}
	return func(t *testing.T) {
		rmcmd := exec.Command("rm", "./werker_files.tar")
		var rmout, rmerr bytes.Buffer
		rmcmd.Stdout = &out
		rmcmd.Stderr = &err
		errr := rmcmd.Run()
		if errr != nil {
			t.Fatal("couldn't clean up werker_files.tar, stdout: ", rmout.String(), "\nstderr: ", rmerr.String(), "\nerror: ", errr.Error())
		}
	}
}

func createDoThingsWebServer() {
	r := mux.NewRouter()
	r.HandleFunc("/do_things.tar", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./werker_files.tar")
	})
	http.ListenAndServe(":3333", r)
}

type Cleanup struct {
	tarCleanup func(t *testing.T)
	dockerCleanup func(t *testing.T)
}

func (c *Cleanup) Clean(t *testing.T) {
	c.tarCleanup(t)
	c.dockerCleanup(t)
}

// CreateLivingDockerContainer will:
//   - create a docker container with an image from imageName
//   - tar up the template directory and serve it on :3333 for testing so the container can download the templates we need
//   - return a cleanup function to defer in your tests
// *assumes you have an internet connection and are running docker on linux or mac*
func CreateLivingDockerContainer(t *testing.T, ctx context.Context, imageName string) (d *Docker, clean func(t *testing.T)) {
	var loopback string
	switch runtime.GOOS {
	case "darwin":
		loopback = "docker.for.mac.localhost"
	case "linux":
		loopback = "172.17.0.1"
	default:
		t.Skip("this test only supports running on darwin or linux")
	}
	b := &Basher{
		LoopbackIp: loopback,
	}
	builder := NewDockerBuilder(b)
	dockerCleaner := &cleaner2.DockerCleaner{}
	d = builder.(*Docker)
	cli, err := client.NewEnvClient()
	if err != nil {
		t.Fatal("couldn't create docker cli, err: ", err.Error())
	}
	d.DockerClient = cli
	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		t.Fatal("couldn't pull image, err: ", err.Error())
	}
	defer out.Close()
	//todo: make sure to call this at some point
	cleanupTar := tarTemplates(t)
	go createDoThingsWebServer()
	containerConfig := &container.Config{
		Image: imageName,
		Cmd: d.DownloadTemplateFiles("3333"),
		AttachStderr: true,
		AttachStdout: true,
		AttachStdin:true,
		Tty:true,
	}
	hostConfig := &container.HostConfig{
		//TODO: have it be overridable via env variable
		Binds: []string{"/var/run/docker.sock:/var/run/docker.sock"},
		//Binds: []string{ homeDirectory + ":/.ocelot", "/var/run/docker.sock:/var/run/docker.sock"},
		NetworkMode: "host",
	}
	resp, err := cli.ContainerCreate(ctx, containerConfig , hostConfig, nil, "")
	if err != nil {
		t.Fatal("could not create container, error: ", err.Error(),"\nwarnings: ", strings.Join(resp.Warnings, "\n  - "))
	}
	d.ContainerId = resp.ID
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		t.Fatal("couldnt create container, error: ", err.Error())
	}
	// i don't know fi we need to create the log, but maybe for errors?
	containerLog, err := cli.ContainerLogs(ctx, d.ContainerId, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow: true,
	})
	if err != nil {
		t.Fatal("couldn't get container log, error: ", err.Error())
	}
	d.Log = containerLog
	dockerCleanup := func(t *testing.T){
		logout := make(chan[]byte, 100)
		dockerCleaner.Cleanup(ctx, d.ContainerId, logout)
	}
	cleaner := &Cleanup{
		tarCleanup: cleanupTar,
		dockerCleanup: dockerCleanup,
	}
	return d, cleaner.Clean
}


package docker

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
	"github.com/shankj3/ocelot/build/basher"
	cleaner2 "github.com/shankj3/ocelot/build/cleaner"
)

func tarTemplates(t *testing.T) func(t *testing.T) {
	//tar -cvf werker_files.tar *
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Dir(filename)
	template := filepath.Join(filepath.Dir(filepath.Dir(dir)), "template")
	werkertar := filepath.Join(filepath.Dir(dir), "werker_files.tar")
	here, _ := ioutil.ReadDir(".")
	fmt.Println(here)
	cmd := exec.Command("/bin/sh", "-c", "tar -cvf "+werkertar+" *")
	cmd.Dir = template
	var out, err bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &err
	errr := cmd.Run()
	t.Log(out.String())
	t.Log(cmd.Dir)
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("no caller???? ")
	}
	t.Log(filepath.Dir(filename))
	if errr != nil {
		t.Error(fmt.Sprintf("unable to tar up template direc, stdout: %s \n stderr: %s \n err: %s", out.String(), err.String(), errr.Error()))
	}
	return func(t *testing.T) {
		errr := os.Remove(path.Join(filepath.Dir(filename), "werker_files.tar"))
		if errr != nil {
			if !os.IsNotExist(errr) {
				t.Error("couldn't clean up werker_files.tar, error: ", errr.Error())
			}
		}
	}
}

func createDoThingsWebServer() {
	r := mux.NewRouter()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("no caller???? ")
	}
	filep := filepath.Dir(path.Dir(filename)) + "/werker_files.tar"
	r.HandleFunc("/do_things.tar", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filep)
	})
	http.ListenAndServe(":3333", r)
}

type Cleanup struct {
	tarCleanup    func(t *testing.T)
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
	if testing.Short() {
		t.Skip("skipping docker container create due to -short being set")
	}
	var loopback string
	switch runtime.GOOS {
	case "darwin":
		loopback = "docker.for.mac.localhost"
	case "linux":
		loopback = "172.17.0.1"
	default:
		t.Skip("this test only supports running on darwin or linux")
	}
	b := &basher.Basher{
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
		Image:        imageName,
		Cmd:          d.DownloadTemplateFiles("3333"),
		AttachStderr: true,
		AttachStdout: true,
		AttachStdin:  true,
		Tty:          true,
	}
	hostConfig := &container.HostConfig{
		//TODO: have it be overridable via env variable
		Binds: []string{"/var/run/docker.sock:/var/run/docker.sock"},
		//Binds: []string{ homeDirectory + ":/.ocelot", "/var/run/docker.sock:/var/run/docker.sock"},
		NetworkMode: "host",
	}
	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, "")
	if err != nil {
		t.Fatal("could not create container, error: ", err.Error(), "\nwarnings: ", strings.Join(resp.Warnings, "\n  - "))
	}
	d.ContainerId = resp.ID
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		t.Fatal("couldnt create container, error: ", err.Error())
	}
	// i don't know fi we need to create the log, but maybe for errors?
	containerLog, err := cli.ContainerLogs(ctx, d.ContainerId, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		t.Fatal("couldn't get container log, error: ", err.Error())
	}
	d.Log = containerLog
	dockerCleanup := func(t *testing.T) {
		logout := make(chan []byte, 100)
		dockerCleaner.Cleanup(ctx, d.ContainerId, logout)
	}
	cleaner := &Cleanup{
		tarCleanup:    cleanupTar,
		dockerCleanup: dockerCleanup,
	}
	return d, cleaner.Clean
}

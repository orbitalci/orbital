package dockrhelper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// for testing
var pulledByApi bool

// RobustImagePull will attempt to use the docker api to pull an image
//	if it is not found via the api, it will attempt to use the docker client to pull the image, as that way will
//  pull in authentication from linux/mac/whatever keychain. The API proved useless in this case, as all that is
//  in the client code and it pulled in way too much garbage code from docker go api
//  anyone can feel free to prove me wrong, but github.com/moby/moby has me at my wits'o end
func RobustImagePull(imageName string) (closer io.ReadCloser, err error) {
	// try pulling the image through the api
	ctx := context.Background()
	clie, err := client.NewEnvClient()
	if err != nil {
		return nil, errors.New("could not connect to docker to check for image validity, **WARNING THIS MEANS YOUR BUILD MIGHT FAIL IN THE SETUP STAGE**")
	}
	var out io.ReadCloser
	out, err = clie.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err == nil {
		pulledByApi = true
		return out, nil
	}

	var outb, errb bytes.Buffer

	pulledByApi = false
	// if couldn't pull image through ui, try just calling docker
	cmd := exec.Command("/bin/sh", "-c", "\"command -v docker\"")
	if err := cmd.Run(); err != nil {
		return ioutil.NopCloser(&errb), errors.New("cannot check for docker pull because docker is not installed on the machine")
	}

	cmd = exec.Command("/bin/sh", "-c", "docker pull " + imageName)
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		return ioutil.NopCloser(&errb), errors.New(fmt.Sprintf("An error has occured while trying to pull for image %s. \nFull Error is %s. ", imageName, outb.String() + "\n" + errb.String()))
	}
	return ioutil.NopCloser(&outb), nil
}

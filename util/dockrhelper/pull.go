package dockrhelper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)
func RobustImagePull(imageName string) error {
	// try pulling the image through the api
	ctx := context.Background()
	clie, err := client.NewEnvClient()
	if err != nil {
		return errors.New("could not connect to docker to check for image validity, **WARNING THIS MEANS YOUR BUILD MIGHT FAIL IN THE SETUP STAGE**")
	}
	_, err = clie.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err == nil {
		return nil
	}
	// if couldn't pull image through ui, try just calling docker
	cmd := exec.Command("/bin/sh", "-c", "command -v docker")
	if err := cmd.Run(); err != nil {
		return errors.New("cannot check for docker pull because docker is not installed on the machine")
	}

	cmd = exec.Command("/bin/sh", "-c", "docker pull " + imageName)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf("An error has occured while trying to pull for image %s. \nFull Error is %s. ", imageName, outb.String() + "\n" + errb.String()))
	}
	return nil
}

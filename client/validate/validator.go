package validate

import (
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"errors"
	"context"
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
)

//contains all validators for commands as recognized by ocelot <command> [args]
type OcelotValidator struct{}

func GetOcelotValidator() *OcelotValidator {
	return &OcelotValidator{}
}

func (ocelotValidator OcelotValidator) ValidateConfig(config *pb.BuildConfig) error {
	if len(config.BuildTool) == 0 {
		return errors.New("BuildTool must be specified")
	}
	if len(config.Stages) == 0 {
		return errors.New("there must be at least one stage listed")
	}

	var ok bool
	for _, stg := range config.Stages {
		if len(stg.Name) == 0 {
			return errors.New("double check your stages, name is a required field")
		}
		if stg.Name == "build" { ok = true }
	}

	if !ok {
		return errors.New("build is a required stage")
	}


	// validate can pull image
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return errors.New("could not connect to docker to check for image validity, **WARNING THIS MEANS YOUR BUILD MIGHT FAIL IN THE SETUP STAGE**")
	}

	_ , err = cli.ImagePull(ctx, config.Image, types.ImagePullOptions{})
	if err != nil {
		return errors.New("an error has occurred while trying to pull for image: " + config.Image + " full error: " + err.Error())
	}

	//since image pull was successful, we want to remove from the client's local docker hub to avoid polluting their local docker repo
	_, _ = cli.ImageRemove(ctx, config.Image, types.ImageRemoveOptions{})
	return nil
}
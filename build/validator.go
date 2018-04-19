package build

import (
	"errors"

	"bitbucket.org/level11consulting/ocelot/common/helpers/dockrhelper"
	"bitbucket.org/level11consulting/ocelot/models/pb"
	"github.com/mitchellh/cli"
)

//contains all validators for commands as recognized by ocelot <command> [args]
type OcelotValidator struct{}

func GetOcelotValidator() *OcelotValidator {
	return &OcelotValidator{}
}

//validates config, takes in an optional cli out
func (ocelotValidator OcelotValidator) ValidateConfig(config *pb.BuildConfig, UI cli.Ui) error {
	if len(config.Image) == 0 {
		return errors.New("uh-oh, there is no image listed inside of your ocelot yaml file")
	}

	if len(config.BuildTool) == 0 {
		return errors.New("BuildTool must be specified")
	}
	if UI != nil {
		UI.Info("BuildTool is specified \u2713" )
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

	if UI != nil {
		UI.Info("Required stage `build` exists \u2713" )
	}


	if UI != nil {
		UI.Info("Connecting to docker to check for image validity..." )
	}
	out, err := dockrhelper.RobustImagePull(config.Image)
	if UI != nil {
		if err != nil {
			UI.Error(config.Image + " does not exist or credentials cannot be found")
		} else  {
			out.Close()
			UI.Info(config.Image + " exists \u2713")
		}
	}
	out.Close()
	return err
}
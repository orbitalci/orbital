package validate

import (
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"errors"
)

//contains all validators for commands as recognized by ocelot <command> [args]
type OcelotValidator struct{}

func GetOcelotValidator() *OcelotValidator {
	return &OcelotValidator{}
}

func (ocelotValidator OcelotValidator) ValidateConfig(config *pb.BuildConfig) error {
	//TODO: check if image exists?
	if len(config.BuildTool) == 0 {
		return errors.New("BuildTool must be specified")
	}
	if len(config.Stages) == 0 {
		return errors.New("there must be at least one stage listed")
	}

	var ok bool
	for _, stg := range config.Stages {
		if stg.Name == "build" { ok = true }
	}

	if !ok {
		return errors.New("build is a required stage")
	}
	return nil
}
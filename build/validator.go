package build

import (
	"errors"

	"github.com/mitchellh/cli"
	"github.com/shankj3/ocelot/common/helpers/dockrhelper"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
)

//contains all validators for commands as recognized by ocelot <command> [args]
type OcelotValidator struct{}

func GetOcelotValidator() *OcelotValidator {
	return &OcelotValidator{}
}

//validates config, takes in an optional cli out
func (ocelotValidator OcelotValidator) ValidateConfig(config *pb.BuildConfig, UI cli.Ui) error {
	var err error
	if config.Image == "" && config.MachineTag == "" {
		return errors.New("uh-oh, there is no image AND no machineTag listed inside of your ocelot yaml file... one of these is required")
	}

	if len(config.BuildTool) == 0 {
		return errors.New("BuildTool must be specified")
	}
	if UI != nil {
		UI.Info("BuildTool is specified " + models.CHECKMARK)
	}
	if len(config.Stages) == 0 {
		return errors.New("there must be at least one stage listed")
	}
	// todo: add in checking if any machines match machinetag 
	if config.Image != "" {
		if UI != nil {
			UI.Info("Connecting to docker to check for image validity...")
		}
		out, err := dockrhelper.RobustImagePull(config.Image)
		defer func(){if out != nil {out.Close()}}()
		if UI != nil {
			if err != nil {
				UI.Error(config.Image + " does not exist or credentials cannot be found")
			} else {
				UI.Info(config.Image + " exists " + models.CHECKMARK)
			}
		}
	}
	return err
}

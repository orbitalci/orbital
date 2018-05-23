package build

import (
	"errors"
	"fmt"

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
func (ov *OcelotValidator) ValidateConfig(config *pb.BuildConfig, UI cli.Ui) error {
	var err error
	if config.Image == "" && config.MachineTag == "" {
		return errors.New("uh-oh, there is no image AND no machineTag listed inside of your ocelot yaml file... one of these is required")
	}
	if config.Image != "" && config.MachineTag != "" {
		return errors.New("you cannot have both image and machineTag. they are mutually exclusive")
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

// ValidateWithBranch runs the normal Build Config validation, and then also checks if the config is still valid in
//   context with a specific branch. i.e. if a build is triggered on branch X but the build conf only allows Y and Z,
//   then this function will return an error.
func (ov *OcelotValidator) ValidateWithBranch(buildConf *pb.BuildConfig, branch string, ui cli.Ui) error {
	err := ov.ValidateConfig(buildConf, ui)
	if err != nil {
		return err
	}
	branchOk, err := BranchRegexOk(branch, buildConf.Branches)
	if err != nil {
		return err
	}
	if !branchOk {
		err = errors.New(fmt.Sprintf("branch %s does not match any branches listed: %v", branch, buildConf.Branches))
	}
	return err
}
package build

import (
	"errors"
	"fmt"
	"io"
	"strings"

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
	writeUIInfo(UI, "BuildTool is specified " + models.CHECKMARK)

	if len(config.Stages) == 0 {
		return errors.New("there must be at least one stage listed")
	}
	// todo: add in checking if any machines match machinetag 
	if config.Image != "" {
		writeUIInfo(UI, "Connecting to docker to check for image validity...")
		var out io.ReadCloser
		out, err = dockrhelper.RobustImagePull(config.Image)
		defer func(){if out != nil {out.Close()}}()
		if err != nil {
			writeUIError(UI, config.Image + " does not exist or credentials cannot be found")
		} else {
			writeUIInfo(UI, config.Image + " exists " + models.CHECKMARK)
		}
	}
	return err
}


//CheckViability will check to see if the the branch given is in the BuildConfig's allowed branches list. This should be separate from the validate function, as the validate failure should be stored in the database. A queue validate failure should not be.
func (ov *OcelotValidator) CheckViability(buildConf *pb.BuildConfig, branch string) error {
	branchOk, err := BranchRegexOk(branch, buildConf.Branches)
	if err != nil {
		return err
	}
	if !branchOk {
		return NoViability(fmt.Sprintf("branch %s not in the acceptable branches list: %s", branch, strings.Join(buildConf.Branches, ", ")))
	}
	return nil
}

// ViableCheckData is a holder for all necessary data to say whether or not to skip the build:
//   - currentBranch and a list of the good branches from the build config
//   - commitList: a list of commits to check if [skip] exists in the
type ViableCheckData struct {
	currentBranch string
	goodBranches []string
	commitList []*pb.Commit
}

// NotViable is an error that means that the build config should not be queued for a build
type NotViable struct {
	branch string
	commits []string
	msg string
}

func (dq *NotViable) Error() string {
	return dq.msg
}

// NoViability will return a NotViable error, signaling it won't be queued and shouldn't be stored
func NoViability(msg string) *NotViable {
	return &NotViable{msg:msg}
}


func writeUIInfo(ui cli.Ui, msg string) {
	if ui != nil {
		ui.Info(msg)
	}
}

func writeUIError(ui cli.Ui, msg string) {
	if ui != nil {
		ui.Error(msg)
	}
}
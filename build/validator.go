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


//ValidateBranchAgainstConf will check to see if the the branch given is in the BuildConfig's allowed branches list. This should be separate from the validate function, as the validate failure should be stored in the database. A queue validate failure should not be.
func ValidateBranchAgainstConf(buildConf *pb.BuildConfig, branch string) error {
	branchOk, err := BranchRegexOk(branch, buildConf.Branches)
	if err != nil {
		return err
	}
	if !branchOk {
		return NoViability(fmt.Sprintf("branch %s not in the acceptable branches list: %s", branch, strings.Join(buildConf.Branches, ", ")))
	}
	return nil
}

// NewViable will return an instantiated Viable struct
func NewViable(currentBranch string, goodBranches []string, commits []*pb.Commit, force bool) *Viable {
	return &Viable{
		currentBranch: currentBranch,
		buildBranches: goodBranches,
		commitList:    commits,
		force: 		   force,
	}
}

// Viable is a holder for all necessary data to say whether or not to skip the build, essentially a pre-queue validator:
//   - currentBranch and a list of the good branches from the build config
//   - commitList: a list of commits to check if skip msgs exist
//   - force: whether to force a build
type Viable struct {
	force 		  bool
	currentBranch string
	buildBranches []string
	commitList    []*pb.Commit
}

func (vcd *Viable) SetBuildBranches(branches []string) {
	vcd.buildBranches = branches
}

// Validate will check that the current branch is in the list of buildable branches, and it will check to make sure
//   that the commit list does not contain [skip ci] or [ci skip]. If either of these are true, then a NotViable error
//   will be returned.
//   if (*Viable).force == true then an error will not be returned, and the build should be queued.
func (vcd *Viable) Validate() error {
	// first check if the force flag has been set, because can just return immediately if so
	if vcd.force {
		return nil
	}
	// next, check if branch has a regex match with any of the buildable branches
	branchOk, err := BranchRegexOk(vcd.currentBranch, vcd.buildBranches)
	if err != nil {
		return err
	}
	if !branchOk {
		return NoViability(fmt.Sprintf("branch %s not in the acceptable branches list: %s", vcd.currentBranch, strings.Join(vcd.buildBranches, ", ")))
	}
	// then, see if the commit list contains any skip messages
	for _, commit := range vcd.commitList {
		for _, skipmsg := range models.SkipMsgs {
			if strings.Contains(commit.Message, skipmsg) {
				return NoViability(fmt.Sprintf("build will not be queued because one of %s was found in the commit with hash %s. the full commit message is %s", strings.Join(models.SkipMsgs, " | "), commit.Hash, commit.Message))
			}
		}
	}
	return nil
}

// NotViable is an error that means that this commit should not be queued for a build
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
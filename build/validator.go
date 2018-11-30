package build

import (
	"fmt"
	"io"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
	"github.com/shankj3/ocelot/common/helpers/dockrhelper"
	"github.com/shankj3/ocelot/common/trigger"
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
	writeUIInfo(UI, "BuildTool is specified %s", models.CHECKMARK)

	// todo: add in checking if any machines match machinetag
	if config.Image != "" {
		writeUIInfo(UI, "Connecting to docker to check for image validity...")
		var out io.ReadCloser
		out, err = dockrhelper.RobustImagePull(config.Image)
		defer func() {
			if out != nil {
				out.Close()
			}
		}()
		if err != nil {
			writeUIError(UI, "%s does not exist or credentials cannot be found", config.Image)
		} else {
			writeUIInfo(UI, "%s exists %s", config.Image, models.CHECKMARK)
		}
	}

	if len(config.Stages) == 0 {
		return errors.New("there must be at least one stage listed")
	}
	for ind, stage := range config.Stages {
		if ind == 0 {
			writeUIInfo(UI, "Validating stages... ")
		}
		writeUIInfo(UI, "  %s", stage.Name)
		if len(stage.Script) == 0 {
			return errors.New("Script for stage " + stage.Name + "should not be empty")
		}
		for ind, triggy := range stage.Triggers {
			if ind == 0 {
				writeUIInfo(UI, "    Validating trigger strings...")
			}
			_, err := trigger.Parse(triggy)
			if err != nil {
				writeUIError(UI, "      - %s %s", triggy, models.FAILED)
				return errors.Wrap(err, "'triggers' conditions must follow spec, this one did not: " + triggy)
			}
			writeUIInfo(UI, "      - %s %s", triggy, models.CHECKMARK)
		}
	}
	for ind, subscription := range config.Subscriptions {
		if ind == 0 {
			writeUIInfo(UI, "Validating subscriptions...")
		}
		writeUIInfo(UI, "  Subscription %d:", ind+1)
		if !pb.EnvSafeRegex.Match([]byte(subscription.Alias)) {
			writeUIError(UI, "    Alias (%s) for subscription is not valid %s", subscription.Alias, models.FAILED)
			return errors.Errorf("Alias for subscription must be environment variable safe, ie it must match the regex pattern %s. Your alias, %s, does not.", pb.ENV_SAFE, subscription.Alias)
		}
		writeUIInfo(UI, "    Alias (%s) for subscription is valid %s", subscription.Alias, models.CHECKMARK)
		writeUIInfo(UI, "    Validating branches...")
		for _, branchmap := range subscription.Branches {
			split := strings.Split(branchmap, ":")
			if len(split) != 2 {
				writeUIError(UI, "      - %s %s", branchmap, models.FAILED)
				return errors.Errorf("Unparseable branch mapping for %s, must be in format {{subscribedToBranch}}:{{subscribingBranch}}", branchmap)
			}
			writeUIInfo(UI, "      - %s %s", branchmap, models.CHECKMARK)
		}
	}
	return err
}

//ValidateViability will check:
//  - the branch given is a regex match for one of the buildBranches
//  - the commits in commits don't have any messages containing special skip commands ([skip ci]/[ci skip])
// This can be overriden with force
// If the validation fails, a NotViable error will be returned. This means that you should not queue the build or track it. its unworthy.
func (ov *OcelotValidator) ValidateViability(branch string, buildBranches []string, commits []*pb.Commit, force bool) error {
	// first check if the force flag has been set, because can just return immediately if so
	if force {
		return nil
	}
	// next, check if branch has a regex match with any of the buildable branches
	branchOk, err := trigger.BranchRegexOk(branch, buildBranches)

	if err != nil {
		return err
	}
	if !branchOk {
		return NoViability(fmt.Sprintf("branch %s not in the acceptable branches list: %s", branch, strings.Join(buildBranches, ", ")))
	}
	if commits == nil {
		return nil
	}
	// then, see if the commit list contains any skip messages
	for _, commit := range commits {
		for _, skipmsg := range models.SkipMsgs {
			if strings.Contains(commit.Message, skipmsg) {
				return NoViability(fmt.Sprintf("build will not be queued because one of %s was found in the commit with hash %s. the full commit message is %s", strings.Join(models.SkipMsgs, " | "), commit.Hash, commit.Message))
			}
		}
	}
	return nil
}

func (ov *OcelotValidator) ValidateBranchAgainstConf(buildConf *pb.BuildConfig, branch string) error {
	branchOk, err := trigger.BranchRegexOk(branch, buildConf.Branches)
	if err != nil {
		return err
	}
	if !branchOk {
		return NoViability(fmt.Sprintf("branch %s not in the acceptable branches list: %s", branch, strings.Join(buildConf.Branches, ", ")))
	}
	return nil
}

// NotViable is an error that means that this commit should not be queued for a build
type NotViable struct {
	branch  string
	commits []string
	msg     string
}

func (dq *NotViable) Error() string {
	return dq.msg
}

// NoViability will return a NotViable error, signaling it won't be queued and shouldn't be stored
func NoViability(msg string) *NotViable {
	return &NotViable{msg: msg}
}

func writeUIInfo(ui cli.Ui, msgFmt string, fmtVars ...interface{}) {
	if ui != nil {
		ui.Info(fmt.Sprintf(msgFmt, fmtVars...))
	}
}

func writeUIError(ui cli.Ui, msgFmt string, fmtVars ...interface{}) {
	if ui != nil {
		ui.Error(fmt.Sprintf(msgFmt, fmtVars...))
	}
}

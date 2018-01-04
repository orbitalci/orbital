package validator


import (
	ocelog "bitbucket.org/level11consulting/go-til/log"
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"strings"
)

//before we build pipeline config for werker, validate and make sure this is good candidate
// - check if commit branch matches with ocelot.yaml branch
// - check if ocelot.yaml has at least one step called build
//TODO: move validator out to its own class and whatnot, that way admin or command line client can use to validate
func ValidateOcelotYml(buildConf *pb.BuildConfig, branch string, warn bool) (errors []string, warnings []string, valid bool) {
	var ok bool
	valid = true
	for _, stg := range buildConf.Stages {
		if stg.Name == "build" { ok = true }
	}
	if !ok {
		errors = append(errors, "your ocelot.yml does not have the required `build` stage")
		valid = false
	}
	// validate can pull image
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		errors = addErrorToErrors(errors, "could not create docker client", err)
	}
	_, err = cli.ImagePull(ctx, buildConf.Image, types.ImagePullOptions{})
	if err != nil {
		errors = addErrorToErrors(errors, "could not pull image " + buildConf.Image, err)
	}
	if warn {
		for ind, stage := range buildConf.GetStages() {
			if stage.Name == "" {
				warnings = append(warnings, fmt.Sprintf("every stage must have a name, stage at index %d does not", ind))
			}
			if len(stage.Script) == 0 {
				warnings = append(warnings, fmt.Sprintf("stage (index: %d, name: %s) does not have any shell commands. it will do nothing.", ind, stage.Name))
			}
		}
	}

	// check if branch is correct .. maybe this should be external
	var branchOk bool
	for _, buildBranch := range buildConf.Branches {
		if buildBranch == branch {
			branchOk = true
		}
	}
	if !branchOk {
		errors = append(errors,
						fmt.Sprintf("current branch %s does not match any branches listed: %s", branch,
						strings.Join(buildConf.Branches, ", ")))
		valid = false
	}
	ocelog.Log().Errorf("build does not match any branches listed: %v", buildConf.Branches)
	return
}

// Formats all validation errors into something client happy
func PrettyValidationErrors(errors []string, warnings []string) string {
	var buffer bytes.Buffer
	if len(errors) != 0 {
		buffer.WriteString("Validation has failed!\n")
		buffer.WriteString(strings.Join(errors, "\n"))
		buffer.WriteString("\n\n")
	} else {
		buffer.WriteString("Validation has passed without errors!\n\n")
	}
	if len(warnings) > 0 {
		buffer.WriteString("Warnings!\n")
		buffer.WriteString(strings.Join(warnings, "\n"))
		buffer.WriteString("\n\n")
	}
	return buffer.String()
}


func addErrorToErrors(errArray []string, msg string, err error) []string {
	errArray = append(errArray, msg + ", Error: " + err.Error())
	ocelog.IncludeErrField(err).Error(msg)
	return errArray
}
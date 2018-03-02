package commandhelper

import (
	"github.com/mitchellh/cli"
	"google.golang.org/grpc/status"
	"fmt"
	"math"
	"strings"
	"os"
	"os/exec"
	"io/ioutil"
	"context"
	"bitbucket.org/level11consulting/ocelot/admin/models"
)


// UIErrFromGrpc will attempt to use grpc status package to parse out message from rpc err.
// if it is unable, it will use the default message and attach the err.Error() text separated by a newline
func UIErrFromGrpc(err error, ui cli.Ui, defaultMsg string) {
	stat, ok := status.FromError(err)
	if !ok {
		ui.Error(defaultMsg + "\nError: " + err.Error())
	} else {
		ui.Error(stat.Message())
	}
}


//prettifyTime takes in time in seconds and returns a pretty string representation of it
func PrettifyTime(timeInSecs float64) string {
	if timeInSecs < 0 {
		return "running"
	}
	var prettyTime []string
	minutes := int(timeInSecs/60)
	if minutes > 0 {
		prettyTime = append(prettyTime, fmt.Sprintf("%v minutes", minutes))
	}
	seconds := int(math.Mod(timeInSecs, 60))
	if len(prettyTime) > 0 {
		prettyTime = append(prettyTime, "and")
	}
	prettyTime = append(prettyTime, fmt.Sprintf("%v seconds", seconds))
	return strings.Join(prettyTime, " ")
}

//FindCurrentHash will attempt to grab a hash based on running git commands - see client/output/output.go for usage
func FindCurrentHash() string {
	var (
		cmdOut []byte
		cmdHash []byte
		err    error
	)

	cmdName := "git"

	getBranch := []string{"rev-parse", "--abbrev-ref",  "HEAD"}
	if cmdOut, err = exec.Command(cmdName, getBranch...).Output(); err != nil {
		fmt.Fprintln(os.Stderr, "There was an error running git rev-parse command to find the current branch: ", err)
	}

	if len(getBranch) > 0 {
		remoteBranch := fmt.Sprintf("origin/%s", string(cmdOut))
		if cmdHash, err = exec.Command(cmdName, "rev-parse", strings.TrimSpace(remoteBranch)).Output(); err != nil {
			fmt.Fprintln(os.Stderr, "There was an error running git rev-parse command to find the most recently pushed commit: ", err)
		}
	}

	sha := strings.TrimSpace(string(cmdHash))
	return sha
}

//UploadSSHKeyFile will upload the ssh key for a vcs account. This is used by buildcredsadd.go and cred's add.go
func UploadSSHKeyFile (ctx context.Context, ui cli.Ui, oceClient models.GuideOcelotClient, acctName, buildType, sshKeyPath string) int {
	sshKey, err := ioutil.ReadFile(sshKeyPath)
	if err != nil {
		ui.Error(fmt.Sprintf("\tCould not read file at %s \nError: %s", sshKeyPath, err.Error()))
		return 1
	}

	_, err = oceClient.SetVCSPrivateKey(ctx, &models.SSHKeyWrapper{
		AcctName: acctName,
		Type: buildType,
		PrivateKey: sshKey,
	})

	if err != nil {
		ui.Error(fmt.Sprintf("\tCould not upload private key at %s \nError: %s", sshKeyPath, err.Error()))
		return 1
	}

	ui.Info(fmt.Sprintf("\tSuccessfully uploaded private key at %s for %s/%s", sshKeyPath, buildType, acctName))
	return 0
}
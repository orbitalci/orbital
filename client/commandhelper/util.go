package commandhelper

import (
	models "bitbucket.org/level11consulting/ocelot/models/pb"
	"github.com/mitchellh/cli"
	"google.golang.org/grpc/status"

	"context"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strings"
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
func PrettifyTime(timeInSecs float64, queued bool) string {
	if queued {
		return "queued for build"
	}
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


//UploadSSHKeyFile will upload the ssh key for a vcs account. This is used by buildcredsadd.go and cred's add.go
func UploadSSHKeyFile (ctx context.Context, ui cli.Ui, oceClient models.GuideOcelotClient, acctName string, buildType models.SubCredType, sshKeyPath string) int {
	sshKey, err := ioutil.ReadFile(sshKeyPath)
	if err != nil {
		ui.Error(fmt.Sprintf("\tCould not read file at %s \nError: %s", sshKeyPath, err.Error()))
		return 1
	}

	_, err = oceClient.SetVCSPrivateKey(ctx, &models.SSHKeyWrapper{
		AcctName: acctName,
		SubType: buildType,
		PrivateKey: sshKey,
	})

	if err != nil {
		ui.Error(fmt.Sprintf("\tCould not upload private key at %s \nError: %s", sshKeyPath, err.Error()))
		return 1
	}

	ui.Info(fmt.Sprintf("\tSuccessfully uploaded private key at %s for %s/%s", sshKeyPath, buildType, acctName))
	return 0
}

//Debuggit will write given message to WARN of a1 GuideOcelotCmd if there is an environment varible "$DEBUGGIT"
// maybe make it a flag later, not really worried about the performance of all the lookups though since its for debugging
func Debuggit(ui cli.Ui, msg string) {
	if _, ok := os.LookupEnv("DEBUGGIT"); ok {
		ui.Warn(msg)
	}
}
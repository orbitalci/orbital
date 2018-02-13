package commandhelper

import (
	"github.com/mitchellh/cli"
	"google.golang.org/grpc/status"
	"fmt"
	"math"
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
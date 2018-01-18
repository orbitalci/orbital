package commandhelper

import (
	"github.com/mitchellh/cli"
	"google.golang.org/grpc/status"
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
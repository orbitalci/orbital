package commandhelper

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/mitchellh/cli"
	protobuf "github.com/shankj3/ocelot/models/pb"
	"google.golang.org/grpc"
)

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansi)

func MaybeStrip(str string, stripAnsi bool) string {
	if stripAnsi {
		return re.ReplaceAllString(str, "")
	}
	return str
}

/*
This gets embedded inside of ocelot clients and performs helper functions common across all clients.
All the helper functions will print things in the UI. Good rule of thumb is if it doesn't print things,
then it doesn't take cli.Ui as a param.
*/
type OcyHelper struct {
	Hash       string
	AcctRepo   string
	Repo       string
	Account    string
	SuppressUI bool
}

func (oh *OcyHelper) DetectRepo(ui cli.Ui) error {
	if oh.Repo == "ERROR" {
		acctRepo, err := FindAcctRepo()
		ui.Info("Flag -repo was not set, detecting account and repository using git commands")
		if err != nil {
			Debuggit(ui, "error!!! "+err.Error())
			ui.Error("flag -repo must be set or you must be in the directory you wish to view a summary of. see --help")
			return err
		}
		oh.AcctRepo = acctRepo
		if err := oh.SplitAndSetAcctRepo(ui); err != nil {
			return err
		}
		ui.Info(fmt.Sprintf("Detected repository %s and account %s", oh.Repo, oh.Account))
	}
	return nil
}

func (oh *OcyHelper) WriteUi(writer func(string), msg string) {
	if oh.SuppressUI {
		return
	}
	writer(msg)
}

// SplitAndSetAcctRepo will split up the AcctRepo field, and write an error to ui if it doesnt meet spec
func (oh *OcyHelper) SplitAndSetAcctRepo(ui cli.Ui) error {
	Debuggit(ui, "splitting and setting acct repo")
	data := strings.Split(oh.AcctRepo, "/")
	if len(data) != 2 {
		oh.WriteUi(ui.Error, "flag -acct-repo must be in the format <account>/<repo>")
		return errors.New("split created an array of len " + string(len(data)))
	}
	oh.Account = data[0]
	oh.Repo = data[1]
	return nil
}

func (oh *OcyHelper) DetectAcctRepo(ui cli.Ui) error {
	if oh.AcctRepo == "ERROR" {
		acctRepo, err := FindAcctRepo()
		ui.Info("Flag -acct-repo was not set, detecting account and repository using git commands")
		if err != nil {
			Debuggit(ui, "error!!! "+err.Error())
			oh.WriteUi(ui.Error, "flag -acct-repo must be in the format <account>/<repo> or you must be in the directory you wish to view a summary of. see --help")
			return err
		}
		ui.Info("Detected: " + acctRepo)
		oh.AcctRepo = acctRepo
	}
	return nil
}

func (oh *OcyHelper) DetectHash(ui cli.Ui) error {
	if oh.Hash == "ERROR" {
		sha := FindCurrentHash()
		if len(sha) > 0 {
			ui.Info(fmt.Sprintf("no -hash flag passed, using detected hash %s", sha))
			oh.Hash = sha
		} else {
			oh.WriteUi(ui.Error, "flag -hash is required or you must be in directory of tracked project, otherwise there is no build to tail")
			return errors.New("hash not detected, flag not passed")
		}
	}
	return nil
}

//handles streaming for grpc clients,
//***THIS ASSUMES THAT THE STREAMING SERVER HAS A FIELD CALLED OUTPUTLINE****
func (oh *OcyHelper) HandleStreaming(ui cli.Ui, stream grpc.ClientStream, stripAnsi bool) error {
	interrupt := make(chan os.Signal, 2)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-interrupt
		ui.Info("received ctl-c, exiting")
		stream.CloseSend()
		os.Exit(1)
	}()
	for {
		Debuggit(ui, "receiving stream")
		resp := new(protobuf.Response)
		err := stream.RecvMsg(resp)

		if err == io.EOF {
			stream.CloseSend()
			return nil
		} else if err != nil {
			UIErrFromGrpc(err, ui, "Error streaming from werker.")
			return err
		}
		ui.Info(MaybeStrip(resp.GetOutputLine(), stripAnsi))
	}
	return nil
}

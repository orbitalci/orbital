package commandhelper

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mitchellh/cli"
	"github.com/shankj3/ocelot/models"
	protobuf "github.com/shankj3/ocelot/models/pb"
	"google.golang.org/grpc"
)

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
	VcsTypeStr string
	VcsType    protobuf.SubCredType
	SuppressUI bool
}

func (oh *OcyHelper) SetGitHelperFlags(flagger models.Flagger) {
	flagger.StringVar(&oh.AcctRepo, "acct-repo", "ERROR", "<account>/<repo>. if not passed, will attempt detect using git commands")
	flagger.StringVar(&oh.Hash, "hash", "ERROR", "git hash. if not passed, will attempt detect using git commands")
	flagger.StringVar(&oh.VcsTypeStr, "vcs-type", "ERROR", fmt.Sprintf("vcs type of <account>/<repo> (%s). if not passed, will attempt detect using git commands", strings.Join(protobuf.CredType_VCS.SubtypesString(), "|")))

}

func (oh *OcyHelper) DetectRepo(ui cli.Ui) error {
	if oh.Repo == "ERROR" {
		acctRepo, _, err := FindAcctRepo()
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

//DetectAcctRepoVcsType will find the git remote origin of the repository in the current directory if it exists. It will
// then use regex to determine if the repo is either github or bitbucket, and what the account and repository names are.
// The happy path of DetectAcctRepoVcsType will end in OcyHelper's AcctRepo and VcsType fields being set. If an error occurs,
// a user-friendly error will be written to the client UI and the original error will be returned.
func (oh *OcyHelper) DetectAcctRepoVcsType(ui cli.Ui) error {
	var err error
	if oh.AcctRepo == "ERROR" || oh.VcsTypeStr == "ERROR" {
		acctRepo, vcsTyp, findErr := FindAcctRepo()
		if oh.AcctRepo == "ERROR" {
			oh.WriteUi(ui.Info, "Flag -acct-repo was not set, detecting account and repository using git commands")
			oh.AcctRepo = acctRepo
			oh.WriteUi(ui.Info, "Detected <account>/<repo> of " + acctRepo)
		}
		if oh.VcsTypeStr == "ERROR" || oh.VcsTypeStr == "" {
			oh.WriteUi(ui.Info, "Flag -vcs-type not set, detecting from git origin url")
			oh.VcsType = vcsTyp
			oh.WriteUi(ui.Info, "Detected vcs type of " + oh.VcsType.String())
		} else {
			oh.VcsType, err = protobuf.VcsTypeStringToSubCredType(oh.VcsTypeStr)
			if err != nil {
				oh.WriteUi(ui.Error, "Unable to convert -vcs-type to VcsType enum. Error: " + err.Error())
				return err
			}
		}
		if findErr != nil {
			oh.WriteUi(ui.Error, "Unable to detect account/repo/vcs-type from git commands, please report and use the flags to get around this error. Error is: " + findErr.Error())
			return findErr
		}

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
func (oh *OcyHelper) HandleStreaming(ui cli.Ui, stream grpc.ClientStream) error {
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
		ui.Info(resp.GetOutputLine())
	}
	return nil
}

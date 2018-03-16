package commandhelper

import (
	"errors"
	"fmt"
	"strings"
)

type OcyHelper struct {
	Hash 	   string
	AcctRepo   string
	Repo       string
	Account    string
	SuppressUI bool
}

func (oh *OcyHelper) DetectRepo(cmd GuideOcelotCmd) error {
	if oh.Repo == "ERROR" {
		acctRepo, err := FindAcctRepo()
		cmd.GetUI().Info("Flag -repo was not set, detecting account and repository using git commands")
		if err != nil {
			Debuggit(cmd, "error!!! " + err.Error())
			cmd.GetUI().Error("flag -repo must be set or you must be in the directory you wish to view a summary of. see --help")
			return err
		}
		oh.AcctRepo = acctRepo
		if err := oh.SplitAndSetAcctRepo(cmd); err != nil {
			return err
		}
		cmd.GetUI().Info(fmt.Sprintf("Detected repository %s and account %s", oh.Repo, oh.Account))
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
func (oh *OcyHelper) SplitAndSetAcctRepo(cmd GuideOcelotCmd) error {
	data := strings.Split(oh.AcctRepo, "/")
	if len(data) != 2 {
		oh.WriteUi(cmd.GetUI().Error, "flag -acct-repo must be in the format <account>/<repo>")
		return errors.New("split created an array of len " + string(len(data)))
	}
	oh.Account = data[0]
	oh.Repo = data[1]
	return nil
}

func (oh *OcyHelper) DetectAcctRepo(cmd GuideOcelotCmd) error {
	if oh.AcctRepo == "ERROR" {
		acctRepo, err := FindAcctRepo()
		cmd.GetUI().Info("Flag -acct-repo was not set, detecting account and repository using git commands")
		if err != nil {
			Debuggit(cmd, "error!!! " + err.Error())
			oh.WriteUi(cmd.GetUI().Error, "flag -acct-repo must be in the format <account>/<repo> or you must be in the directory you wish to view a summary of. see --help")
			return err
		}
		cmd.GetUI().Info("Detected: " + acctRepo)
		oh.AcctRepo = acctRepo
	}
	return nil
}

func (oh *OcyHelper) DetectHash(cmd GuideOcelotCmd) error {
	if oh.Hash == "ERROR" {
		sha := FindCurrentHash()
		if len(sha) > 0 {
			cmd.GetUI().Info(fmt.Sprintf("no -hash flag passed, using detected hash %s", sha))
			oh.Hash = sha
		} else {
			oh.WriteUi(cmd.GetUI().Error, "flag -hash is required or you must be in directory of tracked project, otherwise there is no build to tail")
			return errors.New("hash not detected, flag not passed")
		}
	}
	return nil
}

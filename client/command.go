package client

import (
	"bitbucket.org/level11consulting/ocelot/client/build"
	"bitbucket.org/level11consulting/ocelot/client/creds"
	"bitbucket.org/level11consulting/ocelot/client/creds/vcscreds"
	"bitbucket.org/level11consulting/ocelot/client/creds/vcscreds/add"
	"bitbucket.org/level11consulting/ocelot/client/creds/vcscreds/list"
	"bitbucket.org/level11consulting/ocelot/client/creds/credsadd"
	"bitbucket.org/level11consulting/ocelot/client/creds/credslist"
	"bitbucket.org/level11consulting/ocelot/client/creds/k8s/add"
	"bitbucket.org/level11consulting/ocelot/client/creds/k8s/list"
	"bitbucket.org/level11consulting/ocelot/client/creds/ssh/sshadd"
	"bitbucket.org/level11consulting/ocelot/client/creds/ssh/sshlist"
	"bitbucket.org/level11consulting/ocelot/client/creds/repocreds"
	"bitbucket.org/level11consulting/ocelot/client/creds/repocreds/add"
	"bitbucket.org/level11consulting/ocelot/client/creds/repocreds/list"
	"bitbucket.org/level11consulting/ocelot/client/kill"
	"bitbucket.org/level11consulting/ocelot/client/output"
	"bitbucket.org/level11consulting/ocelot/client/poll/add"
	"bitbucket.org/level11consulting/ocelot/client/poll/delete"
	"bitbucket.org/level11consulting/ocelot/client/poll/list"
	"bitbucket.org/level11consulting/ocelot/client/status"
	"bitbucket.org/level11consulting/ocelot/client/summary"
	"bitbucket.org/level11consulting/ocelot/client/validate"
	"bitbucket.org/level11consulting/ocelot/client/version"
	"bitbucket.org/level11consulting/ocelot/client/watch"
	ocyVersion "bitbucket.org/level11consulting/ocelot/version"
	"github.com/mitchellh/cli"

	"os"
)

var Commands map[string]cli.CommandFactory

func init() {
	base := &cli.BasicUi{Writer: os.Stdout, ErrorWriter: os.Stderr, Reader: os.Stdin}
	ui := &cli.ColoredUi{Ui: base, OutputColor: cli.UiColorNone, InfoColor: cli.UiColorBlue, ErrorColor: cli.UiColorRed, WarnColor: cli.UiColorYellow}
	verHuman := ocyVersion.GetHumanVersion()
	Commands = map[string]cli.CommandFactory{
		"creds":           func() (cli.Command, error) { return creds.New(), nil },
		"creds add":       func() (cli.Command, error) { return credsadd.New(ui), nil },
		"creds list":      func() (cli.Command, error) { return credslist.New(ui), nil },
		"creds vcs":       func() (cli.Command, error) { return vcscreds.New(), nil },
		"creds vcs list":  func() (cli.Command, error) { return buildcredslist.New(ui), nil },
		"creds vcs add":   func() (cli.Command, error) { return buildcredsadd.New(ui), nil },
		"creds ssh list":  func() (cli.Command, error) { return sshlist.New(ui), nil },
		"creds ssh add":   func() (cli.Command, error) { return sshadd.New(ui), nil },
		"creds repo":      func() (cli.Command, error) { return repocreds.New(), nil },
		"creds repo add":  func() (cli.Command, error) { return repocredsadd.New(ui), nil },
		"creds repo list": func() (cli.Command, error) { return repocredslist.New(ui), nil },
		"creds k8s add":   func() (cli.Command, error) { return kubeadd.New(ui), nil },
		"creds k8s list":   func() (cli.Command, error) { return kubelist.New(ui), nil },
		"logs":            func() (cli.Command, error) { return output.New(ui), nil },
		"summary":         func() (cli.Command, error) { return summary.New(ui), nil },
		"validate":        func() (cli.Command, error) { return validate.New(ui), nil },
		"status":          func() (cli.Command, error) { return status.New(ui), nil },
		"watch":           func() (cli.Command, error) { return watch.New(ui), nil },
		"build":           func() (cli.Command, error) { return build.New(ui), nil },
		"poll":            func() (cli.Command, error) { return polladd.New(ui), nil },
		"poll delete":     func() (cli.Command, error) { return polldelete.New(ui), nil },
		"poll list": 	   func() (cli.Command, error) { return polllist.New(ui), nil },
		"kill":			   func() (cli.Command, error) { return kill.New(ui), nil },
		"version": 		   func() (cli.Command, error) { return version.New(ui, verHuman), nil},
	}
}

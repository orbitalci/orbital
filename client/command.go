package client

import (
	"github.com/mitchellh/cli"
	"github.com/shankj3/ocelot/client/build"
	"github.com/shankj3/ocelot/client/creds/apple"
	"github.com/shankj3/ocelot/client/creds/apple/applelist"
	"github.com/shankj3/ocelot/client/creds/k8s"
	"github.com/shankj3/ocelot/client/creds/ssh"

	"github.com/shankj3/ocelot/client/creds"
	//"github.com/shankj3/ocelot/client/creds/credsadd"
	//"github.com/shankj3/ocelot/client/creds/credslist"
	"github.com/shankj3/ocelot/client/creds/apple/appleadd"
	"github.com/shankj3/ocelot/client/creds/k8s/add"
	"github.com/shankj3/ocelot/client/creds/k8s/list"
	"github.com/shankj3/ocelot/client/creds/repocreds"
	"github.com/shankj3/ocelot/client/creds/repocreds/add"
	"github.com/shankj3/ocelot/client/creds/repocreds/list"
	"github.com/shankj3/ocelot/client/creds/ssh/sshadd"
	"github.com/shankj3/ocelot/client/creds/ssh/sshlist"
	"github.com/shankj3/ocelot/client/creds/vcscreds"
	"github.com/shankj3/ocelot/client/creds/vcscreds/add"
	"github.com/shankj3/ocelot/client/creds/vcscreds/list"
	"github.com/shankj3/ocelot/client/kill"
	"github.com/shankj3/ocelot/client/output"
	"github.com/shankj3/ocelot/client/poll/add"
	"github.com/shankj3/ocelot/client/poll/delete"
	"github.com/shankj3/ocelot/client/poll/list"
	"github.com/shankj3/ocelot/client/repos"
	"github.com/shankj3/ocelot/client/repos/list"
	"github.com/shankj3/ocelot/client/status"
	"github.com/shankj3/ocelot/client/summary"
	"github.com/shankj3/ocelot/client/validate"
	"github.com/shankj3/ocelot/client/version"
	"github.com/shankj3/ocelot/client/watch"
	ocyVersion "github.com/shankj3/ocelot/version"

	"os"
)

var Commands map[string]cli.CommandFactory

func init() {
	base := &cli.BasicUi{Writer: os.Stdout, ErrorWriter: os.Stderr, Reader: os.Stdin}
	ui := &cli.ColoredUi{Ui: base, OutputColor: cli.UiColorNone, InfoColor: cli.UiColorBlue, ErrorColor: cli.UiColorRed, WarnColor: cli.UiColorYellow}
	verHuman := ocyVersion.GetHumanVersion()
	Commands = map[string]cli.CommandFactory{
		"creds":            func() (cli.Command, error) { return creds.New(), nil },
		// todo: fix these  functions then add them back in
		//"creds add":        func() (cli.Command, error) { return credsadd.New(ui), nil },
		//"creds list":       func() (cli.Command, error) { return credslist.New(ui), nil },
		"creds vcs":        func() (cli.Command, error) { return vcscreds.New(), nil },
		"creds vcs list":   func() (cli.Command, error) { return buildcredslist.New(ui), nil },
		"creds vcs add":    func() (cli.Command, error) { return buildcredsadd.New(ui), nil },
		"creds ssh": 	    func() (cli.Command, error) { return ssh.New(), nil },
		"creds ssh list":   func() (cli.Command, error) { return sshlist.New(ui), nil },
		"creds ssh add":    func() (cli.Command, error) { return sshadd.New(ui), nil },
		"creds repo":       func() (cli.Command, error) { return repocreds.New(), nil },
		"creds repo add":   func() (cli.Command, error) { return repocredsadd.New(ui), nil },
		"creds repo list":  func() (cli.Command, error) { return repocredslist.New(ui), nil },
		"creds k8s": 	    func() (cli.Command, error) { return k8s.New(), nil },
		"creds k8s add":    func() (cli.Command, error) { return kubeadd.New(ui), nil },
		"creds k8s list":   func() (cli.Command, error) { return kubelist.New(ui), nil },
		"creds apple": 	    func() (cli.Command, error) { return apple.New(), nil },
		"creds apple add":  func() (cli.Command, error) { return appleadd.New(ui), nil},
		"creds apple list": func() (cli.Command, error) { return applelist.New(ui), nil},
		"logs":             func() (cli.Command, error) { return output.New(ui), nil },
		"summary":          func() (cli.Command, error) { return summary.New(ui), nil },
		"validate":         func() (cli.Command, error) { return validate.New(ui), nil },
		"status":           func() (cli.Command, error) { return status.New(ui), nil },
		"watch":            func() (cli.Command, error) { return watch.New(ui), nil },
		"build":            func() (cli.Command, error) { return build.New(ui), nil },
		"poll":             func() (cli.Command, error) { return polladd.New(ui), nil },
		"poll delete":      func() (cli.Command, error) { return polldelete.New(ui), nil },
		"poll list":        func() (cli.Command, error) { return polllist.New(ui), nil },
		"kill":             func() (cli.Command, error) { return kill.New(ui), nil },
		"version":          func() (cli.Command, error) { return version.New(ui, verHuman), nil },
		"repos":            func() (cli.Command, error) { return repos.New(), nil },
		"repos list":       func() (cli.Command, error) { return reposlist.New(ui), nil },
	}
}

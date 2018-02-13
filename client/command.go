package client

import (
	"bitbucket.org/level11consulting/ocelot/client/creds"
	"bitbucket.org/level11consulting/ocelot/client/creds/buildcreds"
	"bitbucket.org/level11consulting/ocelot/client/creds/buildcreds/add"
	"bitbucket.org/level11consulting/ocelot/client/creds/buildcreds/list"
	"bitbucket.org/level11consulting/ocelot/client/creds/credsadd"
	"bitbucket.org/level11consulting/ocelot/client/creds/credslist"
	"bitbucket.org/level11consulting/ocelot/client/creds/repocreds"
	"bitbucket.org/level11consulting/ocelot/client/creds/repocreds/add"
	"bitbucket.org/level11consulting/ocelot/client/creds/repocreds/list"
	"bitbucket.org/level11consulting/ocelot/client/output"
	"bitbucket.org/level11consulting/ocelot/client/summary"
	"github.com/mitchellh/cli"
	"os"
	"bitbucket.org/level11consulting/ocelot/client/validate"
	"bitbucket.org/level11consulting/ocelot/client/status"
)

var Commands map[string]cli.CommandFactory

func init(){
	base := &cli.BasicUi{Writer: os.Stdout, ErrorWriter: os.Stderr, Reader: os.Stdin}
	ui := &cli.ColoredUi{Ui: base, OutputColor: cli.UiColorNone, InfoColor: cli.UiColorBlue, ErrorColor: cli.UiColorRed, WarnColor: cli.UiColorYellow}
	Commands = map[string]cli.CommandFactory{
		"creds"          : func()(cli.Command, error) { return creds.New(), nil},
		"creds add"      : func()(cli.Command, error) { return credsadd.New(ui), nil },
		"creds list"     : func()(cli.Command, error) { return credslist.New(ui), nil },
		"creds vcs"      : func()(cli.Command, error) { return buildcreds.New(), nil },
		"creds vcs list" : func()(cli.Command, error) { return buildcredslist.New(ui), nil },
		"creds vcs add"  : func()(cli.Command, error) { return buildcredsadd.New(ui), nil },
		"creds repo"     : func()(cli.Command, error) { return repocreds.New(), nil },
		"creds repo add" : func()(cli.Command, error) { return repocredsadd.New(ui), nil},
		"creds repo list": func()(cli.Command, error) { return repocredslist.New(ui), nil},
		"logs"			 : func()(cli.Command, error) { return output.New(ui), nil},
		"summary"        : func()(cli.Command, error) { return summary.New(ui), nil},
		"validate"		 : func()(cli.Command, error) { return validate.New(ui), nil},
		"status"		 : func()(cli.Command, error) { return status.New(ui), nil},
	}
}

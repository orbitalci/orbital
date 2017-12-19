package command

import (
	"bitbucket.org/level11consulting/ocelot/admin/command/creds"
	"bitbucket.org/level11consulting/ocelot/admin/command/creds/buildcreds"
	"bitbucket.org/level11consulting/ocelot/admin/command/creds/buildcreds/add"
	"bitbucket.org/level11consulting/ocelot/admin/command/creds/buildcreds/list"
	"bitbucket.org/level11consulting/ocelot/admin/command/creds/repocreds"
	"bitbucket.org/level11consulting/ocelot/admin/command/creds/repocreds/add"
	"bitbucket.org/level11consulting/ocelot/admin/command/creds/repocreds/list"
	"github.com/mitchellh/cli"
	"os"
)

var Commands map[string]cli.CommandFactory

func init(){
	base := &cli.BasicUi{Writer: os.Stdout, ErrorWriter: os.Stderr, Reader: os.Stdin}
	ui := &cli.ColoredUi{Ui: base, OutputColor: cli.UiColorNone, InfoColor: cli.UiColorBlue, ErrorColor: cli.UiColorRed, WarnColor: cli.UiColorYellow}
	Commands = map[string]cli.CommandFactory{
		"creds list"     : func()(cli.Command, error) { return creds.New(ui), nil },
		"creds vcs"      : func()(cli.Command, error) { return buildcreds.New(), nil },
		"creds vcs list" : func()(cli.Command, error) { return buildcredslist.New(ui), nil },
		"creds vcs add"  : func()(cli.Command, error) { return buildcredsadd.New(ui), nil },
		"creds repo"     : func()(cli.Command, error) { return repocreds.New(), nil },
		"creds repo add" : func()(cli.Command, error) { return repocredsadd.New(ui), nil},
		"creds repo list": func()(cli.Command, error) { return repocredslist.New(ui), nil},
	}
}


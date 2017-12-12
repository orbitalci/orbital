package command

import (
	"bitbucket.org/level11consulting/ocelot/admin/command/buildcreds/add"
	"bitbucket.org/level11consulting/ocelot/admin/command/buildcreds/list"
	"github.com/mitchellh/cli"
	"os"
)

var Commands map[string]cli.CommandFactory

func init(){
	ui := &cli.BasicUi{Writer: os.Stdout, ErrorWriter: os.Stderr, Reader: os.Stdin}
	Commands = map[string]cli.CommandFactory{
		"creds list": func()(cli.Command, error) { return buildcredslist.New(ui), nil },
		"creds add" : func()(cli.Command, error) { return buildcredsadd.New(ui), nil},
	}
}


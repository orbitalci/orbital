package command

import (
	"bitbucket.org/level11consulting/ocelot/admin/command/buildcreds"
	"github.com/mitchellh/cli"
	"os"
)

var Commands map[string]cli.CommandFactory

func init(){
	ui := &cli.BasicUi{Writer: os.Stdout, ErrorWriter: os.Stderr, Reader: os.Stdin}
	Commands = map[string]cli.CommandFactory{
		"creds list": func()(cli.Command, error) { return buildcreds.New(ui), nil },
	}
}


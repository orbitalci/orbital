package main

import (
	"bitbucket.org/level11consulting/ocelot/admin/command"
	"fmt"
	"github.com/mitchellh/cli"
	"os"
)


func mainDo() int {
	args := os.Args[1:]

	// todo how to get just keys out of map?
	var cmds []string
	for c := range command.Commands {
		cmds = append(cmds, c)
	}
	clie := &cli.CLI{
		Args: args,
		Commands: command.Commands,
		Autocomplete: true,
		Name: "ocelot",
		Version: "0.1.0",
		HelpFunc: cli.FilteredHelpFunc(cmds, cli.BasicHelpFunc("ocelot")),
	}

	exitCode, err := clie.Run()
	if err != nil {
		fmt.Errorf("wah")
	}
	return exitCode
}


func main(){
	os.Exit(mainDo())
}
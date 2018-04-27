package main

import (
	"fmt"
	"github.com/mitchellh/cli"
	"github.com/shankj3/ocelot/client"
	"os"
)

func mainDo() int {
	args := os.Args[1:]

	// todo how to get just keys out of map?
	var cmds []string
	for c := range client.Commands {
		cmds = append(cmds, c)
	}
	clie := &cli.CLI{
		Args:         args,
		Commands:     client.Commands,
		Autocomplete: true,
		Name:         "ocelot",
		Version:      "0.1.0",
		HelpFunc:     cli.FilteredHelpFunc(cmds, cli.BasicHelpFunc("ocelot")),
	}

	exitCode, err := clie.Run()
	if err != nil {
		fmt.Errorf("wah")
	}
	return exitCode
}

func main() {
	os.Exit(mainDo())
}

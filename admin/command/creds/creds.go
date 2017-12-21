package creds

import (
"github.com/hashicorp/consul/command/flags"
"github.com/mitchellh/cli"
)

func New() *helpcmd {
	return &helpcmd{}
}

type helpcmd struct{}

func (c *helpcmd) Run(args []string) int {
	return cli.RunResultHelp
}

func (c *helpcmd) Synopsis() string {
	return helpcmdSynopsis
}

func (c *helpcmd) Help() string {
	return flags.Usage(helpcmdHelp, nil)
}

const helpcmdSynopsis = "credential configuration"
const helpcmdHelp = `
Usage: ocelot creds <subcommand> [options] [args]
  Interacting with ocelot credentials!
  This command has subcommands for interacting with ocelot's cred stores
  For more examples, ask for subcommand help or view the documentation.
`

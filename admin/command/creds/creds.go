package creds

import (
"github.com/hashicorp/consul/command/flags"
"github.com/mitchellh/cli"
)

func NewCred() *helpcmd {
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

  This command has subcommands for interacting with ocelot's cred stores
  Current Options:
    - ocelot creds repo
    - ocelot creds vcs
  For more examples, ask for subcommand help or view the documentation.
`

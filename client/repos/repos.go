package repos

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

const helpcmdSynopsis = "repository configuration"
const helpcmdHelp = `
Usage: ocelot repos <subcommand> [options] [args]
  Interacting with ocelot tracked repos
  This client has subcommands for interacting with ocelot's tracked repositories
  For more examples, ask for subcommand help or view the documentation.
`

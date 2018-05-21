package apple

import (
	"github.com/hashicorp/consul/command/flags"
	"github.com/mitchellh/cli"
)

func New() *cmd {
	return &cmd{}
}

type cmd struct{}

func (c *cmd) Run(args []string) int {
	return cli.RunResultHelp
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return flags.Usage(help, nil)
}

const synopsis = "apple developer profile credential configuration"
const help = `
Usage: ocelot creds apple <subcommand> [options] [args]

  This command has subcommands for interacting with Ocelot's apple cred store.
  The Apple credential store is what holds the developer profiles for signing and using proprietary apple tools during builds.

  For more examples, ask for subcommand help or view the documentation.
`


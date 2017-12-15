package buildcreds

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

const synopsis = "ocelot vcs configuration"
const help = `
Usage: ocelot creds vcs <subcommand> [options] [args]

  This command has subcommands for interacting with Ocelot's VCS cred store.
  The VCS cred store is what drives what code repositories are watched by Ocelot.

  Create credential set to trigger your account to be watched by ocelot

      $ ocelot creds vcs add

  List VCS accounts that are tracked by Ocelot

      $ ocelot creds vcs list

  For more examples, ask for subcommand help or view the documentation.
`


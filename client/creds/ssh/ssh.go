package ssh

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

const synopsis = "ssh credentials during builds"
const help = `
Usage: ocelot creds ssh <subcommand> [options] [args]

  This command has subcommands for interacting with Ocelot's ssh cred store.
  Ocelot has the ability to store ssh key files for use to scp/ssh to other machines securely during the build process. 

  For more examples, ask for subcommand help or view the documentation.
`


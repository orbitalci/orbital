package env

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

const synopsis = "environment variable secrets"
const help = `
Usage: ocelot creds env <subcommand> [options] [args]

  This command has subcommands for interacting with Ocelot's environment variable cred store.
  Ocelot securely stores environment variables for use during builds. All builds under an account will have access to that account's environment variables.

  For more examples, ask for subcommand help or view the documentation.
`

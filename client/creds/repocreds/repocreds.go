package repocreds

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

const synopsis = "repo credential configuration"
const help = `
Usage: ocelot creds repo <subcommand> [options] [args]

  This command has subcommands for interacting with Ocelot's Repo cred store.
  The Repo cred store is what saves artifact repository intetgrations, ie Nexus or Artifactory.

  Create credential set to trigger your account to be watched by ocelot

      $ ocelot creds repo add

  List Repo accounts that are tracked by Ocelot

      $ ocelot creds repo list

  For more examples, ask for subcommand help or view the documentation.
`

package helmrepocli

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

const synopsis = "helm repository integrations"
const help = `
Usage: ocelot creds helmrepo <subcommand> [options] [args]

  This command has subcommands for interacting with ocelot's environment variable cred store.
  When a new helm repository is added, ocelot will initialize helm for builds and add the helm repo to the client. 

  For more examples, ask for subcommand help or view the documentation.
`


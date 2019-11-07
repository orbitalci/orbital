package k8s

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

const synopsis = "kubernetes kubeconfig credentials"
const help = `
Usage: ocelot creds k8s <subcommand> [options] [args]

  This command has subcommands for interacting with Ocelot's kubernetes cred store.
  Ocelot securely stores your kubeconfig for deployment into a kubernetes cluster. 

  For more examples, ask for subcommand help or view the documentation.
`

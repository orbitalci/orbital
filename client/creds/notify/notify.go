package notify

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

const synopsis = "notifications secrets configuration"
const help = `
Usage: ocelot creds notify <subcommand> [options] [args]

  This command has subcommands for interacting with Ocelot's notify cred store.
  The notify cred store is what ocelot will use to create notifications off of build statuses. The current only supported notify type is slack.
  To add a new notify credential
      $ ocelot creds notify add -acctname level11consulting -identifier SLACK_L11 -url https://slack.notify/url

  List Notify Credentials that are stored by Ocelot

      $ ocelot creds notify list

  For more examples, ask for subcommand help or view the documentation.
`



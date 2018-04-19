package version

import (
	"flag"

	"bitbucket.org/level11consulting/ocelot/version"
	"github.com/mitchellh/cli"
)

func New(ui cli.Ui, version string) *cmd {
	cm := &cmd{UI: ui, version: version}
	cm.init()
	return cm
}

type cmd struct {
	UI      cli.Ui
	version string
	flags   *flag.FlagSet
	short   bool
}

func (c *cmd) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.BoolVar(&c.short, "short", false, "include if you do not want to have it also return the commit you are at locally")
}

func (c *cmd) Run(args []string) int {
	c.flags.Parse(args)
	if c.short {
		c.UI.Output(version.GetShort())
	} else {
		c.UI.Output(version.GetHumanVersion())
	}
	return 0
}

func (c *cmd) Synopsis() string {
	return "Prints the Ocelot version"
}

func (c *cmd) Help() string {
	return ""
}

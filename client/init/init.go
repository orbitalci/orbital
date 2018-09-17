package ocyinit

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mitchellh/cli"
	"github.com/shankj3/ocelot/client/commandhelper"
	models "github.com/shankj3/ocelot/models/pb"
)

const synopsis = "create a skeleton ocelot.yml file"
const help = `
Usage: ocelot init
  Call init to create a skeleton ocelot.yml file. 
  If flag --render-tag is called, it will render the machineTag field instead of the image field, thus assuming a non-docker build.
  If flag --notify is called, a notify section with empty slack fields will be added to the ocelot.yml as well 
`

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui, config: commandhelper.Config}
	c.init()
	return c
}

type cmd struct {
	UI           cli.Ui
	flags        *flag.FlagSet
	config       *commandhelper.ClientConfig
	renderTag    bool
	renderNotify bool
}

func (c *cmd) GetClient() models.GuideOcelotClient {
	return c.config.Client
}

func (c *cmd) GetUI() cli.Ui {
	return c.UI
}

func (c *cmd) GetConfig() *commandhelper.ClientConfig {
	return c.config
}

func (c *cmd) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.BoolVar(&c.renderTag, "render-tag", false, "call to generate a `machineTag` field instead of a `image` field")
	c.flags.BoolVar(&c.renderNotify, "notify", false, "call to generate the `notify` section, will be omitted otherwise")
}

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	var buildtypefield, notify string
	if c.renderNotify {
		notify = notifySkeleton
	}
	if c.renderTag {
		buildtypefield = "machineTag"
	} else {
		buildtypefield = "image"
	}
	yml := "ocelot.yml"
	_, err := os.Stat(yml)
	switch {
	case os.IsNotExist(err): // do nothing, we are good to go
	case err != nil:
		c.UI.Error("An unexpected error occurred trying to check for the prior existence of an ocelot.yml file. \nError is: " + err.Error())
		return 1
	case err == nil:
		c.UI.Error("There is already an ocelot.yml file in this directory, and I'm not going to overwrite it. Please delete the file if you wish to continue with generating a skeleton.")
		return 1
	}
	rendered := fmt.Sprintf(skeleton, buildtypefield, notify)
	err = ioutil.WriteFile(yml, []byte(rendered), 0600)
	if err != nil {
		c.UI.Error("Unable to write file to ocelot.yml in this location, will print to stdout instead. For your edification, the error is: " + err.Error())
		c.UI.Info(rendered)
		return 1
	}
	dir, _ := os.Getwd()
	if dir != "" {
		dir = fmt.Sprintf("\n%s%socelot.yml", dir, string(os.PathSeparator))
	}
	c.UI.Info("Successfully rendered an ocelot.yml file in the current directory" + dir)
	return 0
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return help
}

const notifySkeleton = `
notify:
  slack:
    channel: ""
    identifier: "" 
    on:
      - "FAIL"`

const skeleton = `%s:
buildTool: %s
branches:
  - master
env: []
stages:
  - name: ""
    trigger: 
      branches: []
    script: []
    env: []
`

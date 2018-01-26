package validate

import (
	"github.com/mitchellh/cli"
	"flag"
	"context"
	"io/ioutil"
	"fmt"
	"bitbucket.org/level11consulting/go-til/deserialize"
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"strings"
	"bitbucket.org/level11consulting/ocelot/client/commandhelper"
	"bitbucket.org/level11consulting/ocelot/admin/models"
)

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui, config: commandhelper.Config}
	c.init()
	return c
}

type cmd struct {
	UI      cli.Ui
	flags   *flag.FlagSet
	ocelotFileLoc string
	config *commandhelper.ClientConfig
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
	c.flags = flag.NewFlagSet("", flag.ExitOnError)
}


func (c *cmd) validateOcelotYaml(ctx context.Context) int {
	conf := &pb.BuildConfig{}
	dese := deserialize.New()
	confFile, err := ioutil.ReadFile(c.ocelotFileLoc)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Could not read file at %s\nError: %s", c.ocelotFileLoc, err.Error()))
		return 1
	}

	if err = dese.YAMLToStruct(confFile, conf); err != nil {
		c.UI.Error(fmt.Sprintf("Could not process file, please check make sure the file at %s exists\nError: %s", c.ocelotFileLoc, err.Error()))
		return 1
	}

	fileName := c.ocelotFileLoc[strings.LastIndex(c.ocelotFileLoc, "/") + 1:]
	if fileName != "ocelot.yml" {
		c.UI.Error("Your file must be named ocelot.yml")
		return 1
	}

	fileValidator := GetOcelotValidator()
	err = fileValidator.ValidateConfig(conf, c.UI)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Invalid ocelot.yml file: %s", err.Error()))
		return 1
	}

	c.UI.Info(fmt.Sprintf("%s is valid", c.ocelotFileLoc))
	return 0
}

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		c.UI.Error("an error occurred while parsing flags")
		return 1
	}

	if len(args) == 0 {
		return cli.RunResultHelp
	}

	c.ocelotFileLoc = args[0]
	ctx := context.Background()
	return c.validateOcelotYaml(ctx)
}

func (c *cmd) Synopsis() string {
	return helpcmdSynopsis
}

func (c *cmd) Help() string {
	return helpcmdHelp
}

const helpcmdSynopsis = "built-in validator"
const helpcmdHelp = `
Usage: ocelot validate [options] [args]
  Interacting with ocelot validator
  This client takes in an argument as a path to a local ocelot.yaml file
  Example: ocelot validate /home/mariannef/git/MyProject/ocelot.yml
`

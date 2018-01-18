package validate

import (
	"github.com/hashicorp/consul/command/flags"
	"github.com/mitchellh/cli"
	"bitbucket.org/level11consulting/ocelot/client/commandhelper"
	"flag"
	"context"
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
	config  *commandhelper.ClientConfig
}

func (c *cmd) GetUI() cli.Ui {
	return c.UI
}

func (c *cmd) GetConfig() *commandhelper.ClientConfig {
	return c.config
}

func (c *cmd) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.StringVar(&c.ocelotFileLoc, "ocelotyml-loc", "",
		"Location of ocelot.yml file to validate")
}


func (c *cmd) validateOcelotYaml(ctx context.Context) int {
	//credWrap := &models.CredWrapper{}
	//dese := deserialize.New()
	//confFile, err := ioutil.ReadFile(c.fileloc)
	//if err != nil {
	//	c.UI.Error(fmt.Sprintf("Could not read file at %s \nError: %s", c.fileloc, err.Error()))
	//	return 1
	//}
	//if err = dese.YAMLToProto(confFile, credWrap); err != nil {
	//	c.UI.Error(fmt.Sprintf("Could not process file, please check documentation\nError: %s", err.Error()))
	//	return 1
	//}
	var errOccured bool
	//if len(credWrap.Vcs) == 0 {
	//	c.UI.Error("Did not read any credentials! Is your yaml formatted correctly?")
	//	return 1
	//}
	//for _, configVal := range credWrap.Vcs {
	//	_, err = c.config.Client.SetVCSCreds(ctx, configVal)
	//	if err != nil {
	//		c.UI.Error(fmt.Sprintf("Could not add credentials for account: %s \nError: %s", configVal.AcctName, err.Error()))
	//		errOccured = true
	//	} else {
	//		c.UI.Info(fmt.Sprintf("Added credentials for account: %s", configVal.AcctName))
	//	}
	//}
	if errOccured {
		return 1
	}
	return 0
}

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}

	if c.ocelotFileLoc != "" {
		return c.validateOcelotYaml(ctx)
	} else {
		return cli.RunResultHelp
	}
	return 0
}

func (c *cmd) Synopsis() string {
	return helpcmdSynopsis
}

func (c *cmd) Help() string {
	return flags.Usage(helpcmdHelp, nil)
}

const helpcmdSynopsis = "built-in validator"
const helpcmdHelp = `
Usage: ocelot validate [options] [args]
  Interacting with ocelot validator
  This client takes in an argument as a path to a local ocelot.yaml file
  Example: ocelot validate /home/mariannef/git/MyProject/ocelot.yml
`

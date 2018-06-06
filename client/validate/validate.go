package validate

import (
	"context"
	"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"github.com/shankj3/go-til/deserialize"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/client/commandhelper"
	models "github.com/shankj3/ocelot/models/pb"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui, config: commandhelper.Config}
	c.init()
	return c
}

type cmd struct {
	UI            cli.Ui
	flags         *flag.FlagSet
	ocelotFileLoc string
	branch        string
	config        *commandhelper.ClientConfig
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
	c.flags.StringVar(&c.ocelotFileLoc, "file-loc", "DEP", "*DEPRECATED! now -file-loc is not necessary* location of your ocelot.yml file")
	c.flags.StringVar(&c.branch, "branch", "", "branch to validate ocelot.yml against. if this is passed, then the validator will also check to see if this matches the buildable branches list in the ocelot file. OPTIONAL")
}

func (c *cmd) validateOcelotYaml(ctx context.Context, ocelotFile string) int {
	conf := &models.BuildConfig{}
	dese := deserialize.New()
	confFile, err := ioutil.ReadFile(ocelotFile)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Could not read file at %s\nError: %s", ocelotFile, err.Error()))
		return 1
	}

	if err = dese.YAMLToStruct(confFile, conf); err != nil {
		c.UI.Error(fmt.Sprintf("Could not process file, please check make sure the file at %s exists\nError: %s", ocelotFile, err.Error()))
		return 1
	}

	fileName := ocelotFile[strings.LastIndex(ocelotFile, string(filepath.Separator))+1:]
	if fileName != "ocelot.yml" {
		c.UI.Error("Your file must be named ocelot.yml")
		return 1
	}

	fileValidator := build.GetOcelotValidator()
	err = fileValidator.ValidateConfig(conf, c.UI)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Invalid ocelot.yml file: %s", err.Error()))
		return 1
	}
	if c.branch != "" {
		err := fileValidator.CheckViability(conf, c.branch)
		if err != nil {
			c.UI.Error("This branch would not build, the validation error was: " + err.Error())
			return 1
		}
	}

	c.UI.Info(fmt.Sprintf("%s is valid", ocelotFile))
	return 0
}

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	ocelotFile := ""

	// Check for arg validation
	args = c.flags.Args()
	switch len(args) {
	case 0:
	case 1:
		ocelotFile = args[0]
	default:
		c.UI.Error(fmt.Sprintf("Too many arguments (expected 1, got %d)", len(args)))
		return 1
	}
	// if ocelot not passed the new way, just on the command line without a flag,
	// first check if the flag was passed...
	if c.ocelotFileLoc != "DEP" {
		ocelotFile = c.ocelotFileLoc
	}
	// then check the current directory for an ocelot.yml file...
	// if neither of these works, then give up
	if ocelotFile == "" {
		dir, err := os.Getwd()
		if err != nil {
			c.UI.Error("Unable to get the current working directory, so not able to detect ocelot.yml file in current location. Please pass as an argument.")
			return 1
		}
		_, err = os.Stat(filepath.Join(dir, "ocelot.yml"))
		if err == nil {
			ocelotFile = filepath.Join(dir, "ocelot.yml")
		} else {
			c.UI.Error("Could not find ocelot.yml in current directory. Please pass ocelot.yml location as an argument.")
			return 1
		}
	}

	ctx := context.Background()
	return c.validateOcelotYaml(ctx, ocelotFile)
}

func (c *cmd) Synopsis() string {
	return helpcmdSynopsis
}

func (c *cmd) Help() string {
	return helpcmdHelp
}

const helpcmdSynopsis = "built-in validator"
const helpcmdHelp = `
Usage: ocelot validate /home/mariannef/git/MyProject/ocelot.yml
  Interacting with ocelot validator
  This client takes a positional argument of the ocelot.yml file, it must be last in the command entered:
    Example: ocelot validate /home/mariannef/git/MyProject/ocelot.yml
  If the file location is omitted, the client will attempt to find an ocelot.yml in the current directory and validate that. 
    Example:
      $ pwd
        /home/mariannef/git/MyProject
      $ ocelot validate
        BuildTool is specified ✓
        Connecting to docker to check for image validity...
        maven:3.5.3 exists ✓
        /home/mariannef/git/MyProject/ocelot.yml is valid
  If you pass the -branch flag, then you can test to see if the given branch would result in a build
    Example:
      $ ocelot validate -branch testing ocelot.yml
        BuildTool is specified ✓
        Connecting to docker to check for image validity...
        maven:3.5.3 exists ✓
        This branch would not build, the validation error was: branch testing not in the acceptable branches list: master, release\/.*
`

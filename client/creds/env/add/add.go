package envadd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/shankj3/ocelot/client/commandhelper"
	models "github.com/shankj3/ocelot/models/pb"
	"gopkg.in/yaml.v2"
)

const (
	credExists = `The environment variable with the name %s already exists under the account %s. Do you wish to overwrite? Only a YES will continue and update the environment variable value, otherwise this entry will be skipped.`
	synopsis = "Add a kubeconfig for connection with kubernetes to ocelot"
	// fixme: change help msg
	help = `
Usage: ocelot creds k8s add -acct my_kewl_acct -name cluster_name -kubeconfig=/home/user/.kube/cluster-config.yaml

`)

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui, config: commandhelper.Config}
	c.init()
	return c
}

type cmd struct {
	UI      cli.Ui
	flags   *flag.FlagSet
	fileloc string
	account string
	name    string
	config  *commandhelper.ClientConfig
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
	c.flags.StringVar(&c.fileloc, "envfile", "",
		"Location of environment variable yaml to upload")
	c.flags.StringVar(&c.account, "acct", "ERROR",
		"Account name to file environment variables under")
}


func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return help
}


func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}
	envCred := &models.GenericCreds{SubType:models.SubCredType_ENV}
	if c.account == "ERROR" {
		c.UI.Error("-acct was not provided")
		return 1
	}
	envCred.AcctName = c.account
	if len(c.flags.Args()) == 0 {
		return c.fileUpload()
	} else {
		return c.argUpload()
	}
}


func (c *cmd) fileUpload() int {
	if c.fileloc == "" {
		c.UI.Error("-envfile is required is no environment credentials are passed on the command line")
		return 1
	}
	envs := make(map[string]string)
	yamlFile, err := ioutil.ReadFile(c.fileloc)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Error reading file at %s, please check your filepath. \nError is: \n    %s", c.fileloc, err.Error()))
		return 1
	}
	err = yaml.Unmarshal(yamlFile, envs)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Could not unmarshal yaml file to a map of ENV_NAME: ENV_VALUE pairs. Please read the documentation and check your file at %s.", c.fileloc))
		return 1
	}
	return c.upload(envs)
}

func (c *cmd) argUpload() int {
	var name, value string
	var err error
	envs := make(map[string]string)
	for _, env := range c.flags.Args() {
		name, value, err = splitEnvs(env)
		if err != nil {
			// fixme: make this a clearer error
			c.UI.Error(err.Error())
			return 1
		}
		envs[name] = value
	}
	return c.upload(envs)
}

func (c *cmd) upload(envs map[string]string) int {
	var env *models.GenericCreds
	var ctx context.Context
	credWrap := &models.GenericWrap{}
	ctx = context.Background()
	for identifier, envvalue := range envs {
		env = &models.GenericCreds{AcctName:c.account, Identifier: identifier, SubType: models.SubCredType_ENV, ClientSecret: envvalue}
		exists, err := c.config.Client.GenericCredExists(ctx, env)
		if err != nil {
			// fixme: make this a clearer error
			c.UI.Error(err.Error())
			return 1
		}
		if exists.Exists {
			answer, err := c.UI.Ask(fmt.Sprintf(credExists, identifier, c.account))
			if err != nil {
				// fixme: make this a clearer error
				c.UI.Error(err.Error())
				return 1
			}
			if answer == "YES" {
				if _, err = c.config.Client.UpdateGenericCreds(ctx, env); err != nil {
					// fixme: make this a clearer error
					c.UI.Error(err.Error())
					return 1
				}

			}
		} else {
			credWrap.Creds = append(credWrap.Creds, env)
		}
	}
	if len(credWrap.Creds) > 0 {
		if _, err := c.config.Client.SetGenericCreds(ctx, credWrap); err != nil {
			// fixme: make this a clearer error
			c.UI.Error(err.Error())
			return 1
		}
		c.UI.Info("Successfully uploaded Environment Variables for use in Builds.")
	}
	return 0
}

func splitEnvs(env string) (envName, envValue string, err error) {
	values := strings.SplitN(env, "=", 2)
	if len(values) != 2 {
		fmt.Println(values)
		return "", "", errors.New("Bad environment variable. Must be in format ENV_VAR=ENV_VALUE ")
	}
	envName = values[0]
	envValue = values[1]
	return envName, envValue, nil
}

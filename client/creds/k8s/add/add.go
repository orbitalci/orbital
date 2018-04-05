package kubeadd

import (
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/client/commandhelper"
	"context"
	"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"io/ioutil"
)

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui, config: commandhelper.NewClientConfig()}
	c.init()
	return c
}


type cmd struct {
	UI cli.Ui
	flags   *flag.FlagSet
	fileloc string
	account string
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
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)

	c.flags.StringVar(&c.fileloc, "kubeconfig", "ERROR",
		"Location of kubeconfig file to upload")
	c.flags.StringVar(&c.account, "acct", "ERROR",
		"Account name to file kubeconfig under.")
}


func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}
	k8cred := &models.K8SCreds{}
	if c.account == "ERROR" {
		c.UI.Error("-acct was not provided")
		return 1
	}
	k8cred.AcctName = c.account
	if c.fileloc == "ERROR" {
		c.UI.Error("-kubeconfig required")
		return 1
	}
	kubeconf, err := ioutil.ReadFile(c.fileloc)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Could not read file at %s \nError: %s", c.fileloc, err.Error()))
		return 1
	}
	k8cred.K8SContents = string(kubeconf)

	if _, err = c.config.Client.SetK8SCreds(ctx, k8cred); err != nil {
		c.UI.Error(fmt.Sprintf("Could not add Kubernetes kubeconfig to admin. \nError: %s", err.Error()))
		return 1
	}
	c.UI.Info("Successfully added a kubeconfig to the account " + c.account)
	return 0
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return help
}

const synopsis = "Add a kubeconfig for connection with kubernetes to ocelot"
const help = `
Usage: ocelot creds k8s add -acct my_kewl_acct -kubeconfig=/home/user/kubeconfig.yaml

`

package helmrepoadd

import (
	"context"
	"flag"
	"fmt"

	"github.com/mitchellh/cli"
	"github.com/shankj3/ocelot/client/commandhelper"
	models "github.com/shankj3/ocelot/models/pb"
	"google.golang.org/grpc/status"
)

const (
	credExists = `The helm repo address with the name %s already exists under the account %s. Do you wish to overwrite? Only a YES will continue and update the helm chart repo value, otherwise this entry will be skipped.`
	synopsis   = "Add a helm repository address for downloading charts. "
	// fixme: change help msg
	help = `
Usage: ocelot creds helmrepo add -acct my_kewl_acct -repo-name shankj3_charts -helm-url https://github.io/shankj3_helm_repository
`
)

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui, config: commandhelper.Config}
	c.init()
	return c
}

type cmd struct {
	UI           cli.Ui
	flags        *flag.FlagSet
	chartRepoUrl string
	account      string
	name         string
	repoName     string
	config       *commandhelper.ClientConfig
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
	c.flags.StringVar(&c.account, "acct", "ERROR", "Account name to file environment variables under")
	c.flags.StringVar(&c.repoName, "repo-name", "ERROR", "Identifier for helm chart repo credential. During the build, the helm repo will be added under this name.")
	c.flags.StringVar(&c.chartRepoUrl, "helm-url", "ERROR", "URL of helm chart repository")
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

	if c.account == "ERROR" {
		c.UI.Error("-acct was not provided")
		return 1
	}
	if c.chartRepoUrl == "ERROR" {
		c.UI.Error("-helm-url was not provided")
		return 1
	}
	if c.repoName == "ERROR" {
		c.UI.Error("-repo-name was not provided")
		return 1
	}
	helmCred := &models.GenericCreds{
		SubType:      models.SubCredType_HELM_REPO,
		AcctName:     c.account,
		ClientSecret: c.chartRepoUrl,
		Identifier:   c.repoName,
	}
	return c.upload(helmCred)
}

func getErrMsg(err error) (msg string) {
	statErr, ok := status.FromError(err)
	if !ok {
		msg = err.Error()
	} else {
		msg = statErr.Message()
	}
	return

}

func (c *cmd) upload(helmCred *models.GenericCreds) int {
	var ctx context.Context

	ctx = context.Background()

	exists, err := c.config.Client.GenericCredExists(ctx, helmCred)
	if err != nil {
		c.UI.Error("Unable to check if credential exists, error is: " + getErrMsg(err))
		return 1
	}
	if exists.Exists {
		answer, err := c.UI.Ask(fmt.Sprintf(credExists, helmCred.Identifier, c.account))
		if err != nil {
			c.UI.Error("exiting from input ask, error is " + err.Error())
			return 1
		}
		if answer == "YES" {
			if _, err = c.config.Client.UpdateGenericCreds(ctx, helmCred); err != nil {
				c.UI.Error("unable to update credential, error is: " + getErrMsg(err))
				return 1
			}

		}
	} else {
		if _, err := c.config.Client.SetGenericCreds(ctx, helmCred); err != nil {
			c.UI.Error("Unable to set credential, error is: " + getErrMsg(err))
			return 1
		}
		c.UI.Info(fmt.Sprintf("Uploaded %s", helmCred.Identifier))
	}

	return 0
}

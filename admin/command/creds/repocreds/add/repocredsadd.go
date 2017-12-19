package repocredsadd

import (
	"bitbucket.org/level11consulting/go-til/deserialize"
	"bitbucket.org/level11consulting/ocelot/admin"
	"bitbucket.org/level11consulting/ocelot/admin/command/commandhelper"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"context"
	"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"io/ioutil"
)

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui}
	c.init()
	return c
}


type cmd struct {
	UI cli.Ui
	flags   *flag.FlagSet
	fileloc string
	client  models.GuideOcelotClient
	config *admin.ClientConfig
}

func (c *cmd) GetClient() models.GuideOcelotClient {
	return c.client
}

func (c *cmd) GetUI() cli.Ui {
	return c.UI
}

func (c *cmd) GetConfig() *admin.ClientConfig {
	return c.config
}

func (c *cmd) init() {
	var err error
	//todo: THIS IS HARDCODED! BAD!
	config := admin.NewClientConfig()
	c.client, err = admin.GetClient(config.AdminLocation)
	if err != nil {
		panic(err)
	}
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.StringVar(&c.fileloc, "credfile-loc", "",
		"Location of yaml file containing repo creds to upload")
}

func (c *cmd) runCredFileUpload(ctx context.Context) int {
	credWrap := &models.RepoCredWrapper{}
	dese := deserialize.New()
	confFile, err := ioutil.ReadFile(c.fileloc)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Could not read file at %s \nError: %s", c.fileloc, err.Error()))
		return 1
	}
	if err = dese.YAMLToProto(confFile, credWrap); err != nil {
		c.UI.Error(fmt.Sprintf("Could not process file, please check documentation\nError: %s", err.Error()))
		return 1
	}
	var errOccured bool
	if len(credWrap.Credentials) == 0 {
		c.UI.Error("Did not read any repo credentials! Is your yaml formatted correctly?")
		return 1
	}
	for _, configVal := range credWrap.Credentials {
		_, err = c.client.SetRepoCreds(ctx, configVal)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Could not add credentials for repository: %s \nError: %s", configVal.AcctName, err.Error()))
			errOccured = true
		} else {
			c.UI.Info(fmt.Sprintf("Added credentials for account with username %s: %s", configVal.Username, configVal.AcctName))
		}
	}
	if errOccured {
		return 1
	}
	return 0
}

// seems really unlikely that hashicorps tool will fail, but this way if it does its all in one
// function.
func getCredentialsFromUiAsk(UI cli.Ui) (creds *models.RepoCreds, errorConcat string) {
	creds = &models.RepoCreds{}
	var err error
	if creds.Username, err = UI.Ask("Username: "); err != nil {
		errorConcat += "\n" + "Username Err: " +  err.Error()
	}
	if creds.Type, err = UI.Ask("Type (nexus or artifactory): "); err != nil {
		errorConcat += "\n" + "Type Err: " +  err.Error()
	}
	if creds.AcctName, err = UI.Ask("Account Name: "); err != nil {
		errorConcat += "\n" + "Account Name Err: " +  err.Error()
	}
	if creds.RepoUrl, err = UI.Ask("Repo URL for uploading repo artifacts: "); err != nil {
		errorConcat += "\n" + "Repo URL Err: " + err.Error()
	}
	if creds.Password, err = UI.AskSecret("Password for Repo Integration: "); err != nil {
		errorConcat += "\n" + "Password Err: " + err.Error()
	}
	return creds, errorConcat
}

func (c *cmd) runStdinUpload(ctx context.Context) int {
	creds, errConcat := getCredentialsFromUiAsk(c.UI)
	if errConcat != "" {
		c.UI.Error(fmt.Sprint("Error recieving input: ", errConcat))
		return 1
	}
	if _, err := c.client.SetRepoCreds(ctx, creds); err != nil {
		c.UI.Error(fmt.Sprintf("Could not add credentials for account: %s \nError: %s", creds.AcctName, err.Error()))
		return 1
	}
	c.UI.Info(fmt.Sprint("Successfully added credentials for account: ", creds.AcctName))
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
	if c.fileloc != "" {
		return c.runCredFileUpload(ctx)
	} else {
		return c.runStdinUpload(ctx)
	}
	return 0
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return help
}

const synopsis = "Add credentials or a set of them for artifact repositories"
const help = `
Usage: ocelot creds repo add
  Add one set of credentials or a list of them.
  If you specify a filename using:
    ocelot creds add -credfile-loc=<yaml file>
  The client will expect that the yaml is a repo credentials object with an array of artifact repository creds.
  For example:
    credentials:
    - username: <ARTIFACT_USER>     ---> username for logging into artifact repo (i.e. artifactory / nexus)
      password: <PASSWORD>          ---> password for logging into artifact repo
      repoUrl: <REPO_URL>           ---> e.g. !!!! get url example from marianne !!!!
      acctName: <ACCOUNT_NAME>      ---> e.g. level11consulting
      type: <REPO_TYPE>             ---> e.g. nexus

  Retrieves all credentials that ocelot uses to integrate with artifact repositories
`

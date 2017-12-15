package buildcredsadd


import (
	"bitbucket.org/level11consulting/go-til/deserialize"
	"bitbucket.org/level11consulting/ocelot/admin"
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
}

func (c *cmd) init() {
	var err error
	//todo: THIS IS HARDCODED! BAD!
	c.client, err = admin.GetClient("localhost:10000")
	if err != nil {
		panic(err)
	}
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.StringVar(&c.fileloc, "credfile-loc", "",
		"Location of yaml file containing creds to upload")
}

func (c *cmd) runCredFileUpload(ctx context.Context) int {
	credWrap := &models.CredWrapper{}
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
		c.UI.Error("Did not read any credentials! Is your yaml formatted correctly?")
		return 1
	}
	for _, configVal := range credWrap.Credentials {
		_, err = c.client.SetCreds(ctx, configVal)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Could not add credentials for account: %s \nError: %s", configVal.AcctName, err.Error()))
			errOccured = true
		} else {
			c.UI.Info(fmt.Sprintf("Added credentials for account: %s", configVal.AcctName))
		}
	}
	if errOccured {
		return 1
	}
	return 0
}

// seems really unlikely that hashicorps tool will fail, but this way if it does its all in one
// function.
func getCredentialsFromUiAsk(UI cli.Ui) (creds *models.Credentials, errorConcat string) {
	creds = &models.Credentials{}
	var err error
	if creds.ClientId, err = UI.Ask("Client ID: "); err != nil {
		errorConcat += "\n" + "Client ID Err: " +  err.Error()
	}
	if creds.Type, err = UI.Ask("Type: "); err != nil {
		errorConcat += "\n" + "Type Err: " +  err.Error()
	}
	if creds.AcctName, err = UI.Ask("Account Name: "); err != nil {
		errorConcat += "\n" + "Account Name Err: " +  err.Error()
	}
	if creds.TokenURL, err = UI.Ask("OAuth token URL for repository: "); err != nil {
		errorConcat += "\n" + "OAuth token URL for repository: " + err.Error()
	}
	if creds.ClientSecret, err = UI.AskSecret("Secret for OAuth: "); err != nil {
		errorConcat += "\n" + "OAuth Secret Err: " + err.Error()
	}
	return creds, errorConcat
}

func (c *cmd) runStdinUpload(ctx context.Context) int {
	creds, errConcat := getCredentialsFromUiAsk(c.UI)
	if errConcat != "" {
		c.UI.Error(fmt.Sprint("Error recieving input: ", errConcat))
		return 1
	}
	if _, err := c.client.SetCreds(ctx, creds); err != nil {
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

const synopsis = "Add credentials or a set of them"
const help = `
Usage: ocelot creds add
  Add one set of credentials or a list of them.
  If you specify a filename using:
    ocelot creds add -credfile-loc=<yaml file>
  The client will expect that the yaml is a credentials object with an array of creds you would like to integrate with ocelot.
  For example:
    credentials:
    - clientId: <OAUTH_CLIENT_ID>   ---> client id correlated with clientSecret
      clientSecret: <OAUTH_SECRET>  ---> generated when you add an oauth access
      tokenURL: <OAUTH_TOKEN_URL>   ---> e.g. https://bitbucket.org/site/oauth2/access_token
      acctName: <ACCOUNT_NAME>      ---> e.g. level11consulting
      type: <SCM_TYPE>              ---> e.g. bitbucket

  Retrieves all credentials that ocelot uses to track repositories
`

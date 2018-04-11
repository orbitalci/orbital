package buildcredsadd


import (
	"bitbucket.org/level11consulting/go-til/deserialize"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/client/commandhelper"
	"context"
	"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"io/ioutil"
	"strings"
)

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui, config: commandhelper.Config}
	c.init()
	return c
}


type cmd struct {
	UI      cli.Ui
	flags   *flag.FlagSet
	fileloc string
	sshKeyFile string
	acctName string
	buildType string
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
	c.flags.StringVar(&c.sshKeyFile, "sshfile-loc", "", "location of ssh private key to upload")
	c.flags.StringVar(&c.acctName, "acctname", "", "account name matching with sshfile-loc")
	c.flags.StringVar(&c.buildType, "type", "", "build type for this sshfile. Ex: bitbucket")

	c.flags.StringVar(&c.fileloc, "credfile-loc", "","Location of yaml file containing creds to upload")
}

func (c *cmd) runCredFileUpload(ctx context.Context) int {
	credWrap := &models.CredWrapper{}
	dese := deserialize.New()
	confFile, err := ioutil.ReadFile(c.fileloc)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Could not read file at %s \nError: %s", c.fileloc, err.Error()))
		return 1
	}
	if err = dese.YAMLToStruct(confFile, credWrap); err != nil {
		c.UI.Error(fmt.Sprintf("Could not process file, please check documentation\nError: %s", err.Error()))
		return 1
	}
	for _, cred := range credWrap.Vcs {
		cred.Type = models.CredType_VCS
	}
	var errOccured bool
	if len(credWrap.Vcs) == 0 {
		c.UI.Error("Did not read any credentials! Is your yaml formatted correctly?")
		return 1
	}
	for _, configVal := range credWrap.Vcs {
		_, err = c.config.Client.SetVCSCreds(ctx, configVal)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Could not add credentials for account: %s \nError: %s", configVal.AcctName, err.Error()))
			errOccured = true
		} else {
			c.UI.Info(fmt.Sprintf("Added credentials for account: %s", configVal.AcctName))

			//after creds are successfully uploaded via file, upload ssh key file
			if len(configVal.SshFileLoc) > 0 {
				c.UI.Info(fmt.Sprintf("\tdetected ssh file location: %s", configVal.SshFileLoc))
				commandhelper.UploadSSHKeyFile(ctx, c.UI, c.config.Client, configVal.AcctName, configVal.SubType, configVal.SshFileLoc)
			}
		}
	}
	if errOccured {
		return 1
	}
	return 0
}

// seems really unlikely that hashicorps tool will fail, but this way if it does its all in one
// function.
func getCredentialsFromUiAsk(UI cli.Ui) (creds *models.VCSCreds, errorConcat string) {
	creds = &models.VCSCreds{}
	var err error
	if creds.ClientId, err = UI.Ask("Client ID: "); err != nil {
		errorConcat += "\n" + "Client ID Err: " +  err.Error()
	}
	//if creds.Type, err = UI.Ask("Type: "); err != nil {
	//	errorConcat += "\n" + "Type Err: " +  err.Error()
	//}
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
	if _, err := c.config.Client.SetVCSCreds(ctx, creds); err != nil {
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
	} else if c.acctName == "" && c.sshKeyFile == "" && c.buildType == "" {
		return c.runStdinUpload(ctx)
	}

	if c.acctName != "" && c.sshKeyFile != "" && c.buildType != "" {
		subType, ok:= models.SubCredType_value[strings.ToUpper(c.buildType)]
		if !ok {
			c.UI.Error("-type must be vcs type, ie bitbucket")
		}
		return commandhelper.UploadSSHKeyFile(ctx, c.UI, c.config.Client, c.acctName, models.SubCredType(subType), c.sshKeyFile)
	} else {
		c.UI.Error("-acctname, -sshfile-loc and -type must be passed together, -acctname should correspond with the account you'd like the ssh key file to be associated with, and -type should correspond with your acctname")
		return 1
	}
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return help
}

const synopsis = "Add credentials or a set of them"
const help = `
Usage: ocelot creds vcs add
  Add one set of credentials or a list of them. Credentials may also need an SSH key file, if an ssh key path is populated, the file will be uploaded to vault and associated with the specified account/type. 
  You can specify a filename using:
    ocelot creds add vcs -credfile-loc=<yaml file>
  The client will expect that the yaml is a credentials object with an array of creds you would like to integrate with ocelot.
  For example:
    credentials:
    - clientId: <OAUTH_CLIENT_ID>     ---> client id correlated with clientSecret
      clientSecret: <OAUTH_SECRET>    ---> generated when you add an oauth access
      tokenURL: <OAUTH_TOKEN_URL>     ---> e.g. https://bitbucket.org/site/oauth2/access_token
      acctName: <ACCOUNT_NAME>        ---> e.g. level11consulting
      type: <SCM_TYPE>                ---> e.g. bitbucket
	  sshFileLoc: <SSH_KEY_FILE_PATH> ---> e.g. /home/mariannefeng/.ssh/id_rsa
	
  sshFileLoc is not a required field, if you'd to add it later after a vcs account has already been added to ocelot, it can be uploaded via CLI like so:
    ocelot creds add vcs -acctname <ACCOUNT_NAME> -sshfile-loc <SSH_KEY_FILE_PATH> -type <SCM_TYPE>

  Adds credentials for ocelot to use in tracking repositories
`

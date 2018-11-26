package buildcredsadd

import (
	"context"
	"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
	"github.com/shankj3/go-til/deserialize"
	"github.com/shankj3/ocelot/client/commandhelper"
	models "github.com/shankj3/ocelot/models/pb"
	"io/ioutil"
	"strings"
)

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui, config: commandhelper.Config}
	c.init()
	return c
}

type cmd struct {
	UI         cli.Ui
	flags      *flag.FlagSet
	fileloc    string
	sshKeyFile string
	acctName   string
	buildType  string
	config     *commandhelper.ClientConfig
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

	c.flags.StringVar(&c.fileloc, "credfile-loc", "", "Location of yaml file containing creds to upload")
}

//TODO: fix - this doesn't work - yes it does?
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
	if len(credWrap.Vcs) == 0 {
		c.UI.Error("Did not read any credentials! Is your yaml formatted correctly?")
		return 1
	}
	for _, configVal := range credWrap.Vcs {
		err = uploadCredential(ctx, c.config.Client, c.UI, configVal)
		if err != nil {
			if _, ok := err.(*commandhelper.DontOverwrite); ok {
				return 0
			}
			c.UI.Error(fmt.Sprintf("Could not add credentials for account: %s \nError: %s", configVal.AcctName, err.Error()))
			return 1
		} else {
			c.UI.Info(fmt.Sprintf("Added credentials for account: %s", configVal.AcctName))

			//after creds are successfully uploaded via file, upload ssh key file
			if len(configVal.SshFileLoc) > 0 {
				c.UI.Info(fmt.Sprintf("\tdetected ssh file location: %s", configVal.SshFileLoc))
				commandhelper.UploadSSHKeyFile(ctx, c.UI, c.config.Client, configVal.AcctName, configVal.SubType, configVal.SshFileLoc)
			}
		}
	}
	return 0
}

// uploadCredential will check if credential already exists. if it does, it will ask if the user wishes to overwrite. if the user responds YES, the credential will be updated.
// if it does not exist, will be inserted as normal.
func uploadCredential(ctx context.Context, client models.GuideOcelotClient, UI cli.Ui, cred *models.VCSCreds) error {
	exists, err := client.VCSCredExists(ctx, cred)
	if err != nil {
		return err
	}

	if exists.Exists {
		update, err := UI.Ask(fmt.Sprintf("Entry with Account Name %s and Vcs Type %s already exists. Do you want to overwrite? "+
			"Only a YES will continue with update, otherwise the client will exit. ", cred.AcctName, strings.ToLower(cred.SubType.String())))
		if err != nil {
			return err
		}
		if update != "YES" {
			UI.Info("Did not recieve a YES at the prompt, will not overwrite. Exiting.")
			return &commandhelper.DontOverwrite{}
		}
		_, err = client.UpdateVCSCreds(ctx, cred)
		if err != nil {
			return err
		}
		UI.Error("Succesfully update VCS Credential.")
		return nil
	}
	_, err = client.SetVCSCreds(ctx, cred)
	return err
}

func getCredentialsFromStdin(UI cli.Ui) (creds *models.VCSCreds, err error) {
    creds = &models.VCSCreds{}
	if creds.AcctName, err = UI.Ask("Account Name: "); err != nil {
		return
	}
	var unCastedSt string
	if unCastedSt, err = UI.Ask("Type: "); err != nil {
		return
	}
	if int32SubType, ok := models.SubCredType_value[strings.ToUpper(strings.Replace(unCastedSt, " ", "", -1))]; !ok {
		err = errors.New("Type must be bitbucket|github")
		return
	} else {
		creds.SubType = models.SubCredType(int32SubType)
	}
	switch creds.SubType {
	case models.SubCredType_BITBUCKET:
		if creds.ClientId, err = UI.Ask("Client Key: "); err != nil {
			return
		}
		if creds.ClientSecret, err = UI.AskSecret("Client Secret: "); err != nil {
			return nil, err
		}
	case models.SubCredType_GITHUB:
		creds.ClientId = "UNNECESSARY"
		if creds.ClientSecret, err = UI.Ask("Personal Access Token: "); err != nil {
			return
		}
	}
	return

}

func (c *cmd) runStdinUpload(ctx context.Context) int {
	creds, err := getCredentialsFromStdin(c.UI)
	if err != nil {
		c.UI.Error(fmt.Sprint("Error recieving input: ", err.Error()))
		return 1
	}
	if err := uploadCredential(ctx, c.config.Client, c.UI, creds); err != nil {
		if _, ok := err.(*commandhelper.DontOverwrite); ok {
			return 0
		}
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
		subType, ok := models.SubCredType_value[strings.ToUpper(c.buildType)]
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
  Detailed help for setup: https://github.com/shankj3/ocelot/wiki/client#initial-configuration
  Add one set of credentials or a list of them. Credentials may also need an SSH key file, if an ssh key path is populated, the file will be uploaded to vault and associated with the specified account/type. 
  You can specify a filename using:
    ocelot creds add vcs -credfile-loc=<yaml file>
  The client will expect that the yaml is a credentials object with an array of creds you would like to integrate with ocelot.
  For example:
    credentials:
    - clientId: <OAUTH_CLIENT_ID>          ---> client id correlated with clientSecret
      clientSecret: <OAUTH_SECRET>         ---> generated when you add an oauth access
      acctName: <ACCOUNT_NAME>             ---> e.g. level11consulting
      type: bitbucket
	  sshFileLoc: <SSH_KEY_FILE_PATH>      ---> e.g. /home/mariannefeng/.ssh/id_rsa

    - accessToken: <PERSONAL_ACCESS_TOKEN> ---> personal access token created in github
      acctName: <ACCOUNT_NAME>             ---> e.g. level11consulting
      type: github
      sshFileLoc: <SSH_KEY_FILE_PATH>      ---> e.g. /home/mariannefeng/.ssh/id_rsa
	
  sshFileLoc is not a required field, if you'd to add it later after a vcs account has already been added to ocelot, it can be uploaded via CLI like so:
    ocelot creds add vcs -acctname <ACCOUNT_NAME> -sshfile-loc <SSH_KEY_FILE_PATH> -type <SCM_TYPE>

  Bitbucket
  ---------
  Bitbucket will require creating an OAuth consumer in settings. When that is created, save the clientId and clientSecret and use those values to create this vcs credential. 
  
  Github
  ------
  With Github, instead of creating an OAuth consumer, you must create a personal access token for your account.
  https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
  a ClientId will not be required, only this token. 
`

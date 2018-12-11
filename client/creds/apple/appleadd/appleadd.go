package appleadd

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/level11consulting/ocelot/client/commandhelper"
	models "github.com/level11consulting/ocelot/models/pb"
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
	account    string
	identifier string
	profPw     string
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

	c.flags.StringVar(&c.fileloc, "zip", "ERROR", "Location of export Xcode profile zip to upload")
	c.flags.StringVar(&c.account, "acct", "ERROR", "Account name to file the xcode profile under.")
	c.flags.StringVar(&c.identifier, "identifier", "ERROR", "unique identifier for this ssh key")
	c.flags.StringVar(&c.profPw, "password", "ERROR", "password set when developer profile was exported from xcode")
}

// uploadCredential will check if credential already exists. if it does, it will ask if the user wishes to overwrite. if the user responds YES, the credential will be updated.
// if it does not exist, will be inserted as normal.
func uploadCredential(ctx context.Context, client models.GuideOcelotClient, UI cli.Ui, cred *models.AppleCreds) error {
	exists, err := client.AppleCredExists(ctx, cred)
	if err != nil {
		return err
	}

	if exists.Exists {
		update, err := UI.Ask(fmt.Sprintf("Entry with Account Name %s and Repo Type %s already exists. Do you want to overwrite? "+
			"Only a YES will continue with update, otherwise the client will exit. ", cred.AcctName, strings.ToLower(cred.SubType.String())))
		if err != nil {
			return err
		}
		if update != "YES" {
			UI.Info("Did not recieve a YES at the prompt, will not overwrite. Exiting.")
			return &commandhelper.DontOverwrite{}
		}
		_, err = client.UpdateAppleCreds(ctx, cred)
		if err != nil {
			return err
		}
		UI.Error("Succesfully updated Apple Credential.")
		return nil
	}
	_, err = client.SetAppleCreds(ctx, cred)
	return err
}

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}
	creds := &models.AppleCreds{SubType: models.SubCredType_DEVPROFILE}
	if c.account == "ERROR" {
		c.UI.Error("-acct was not provided")
		return 1
	}
	creds.AcctName = c.account
	if c.identifier == "ERROR" {
		c.UI.Error("-identifier required")
		return 1
	}
	if c.fileloc == "ERROR" {
		c.UI.Error("-zip required")
		return 1
	}
	if c.profPw == "ERROR" {
		var err error
		c.profPw, err = c.UI.AskSecret("Profile Password: ")
		if err != nil {
			c.UI.Error("Error: " + err.Error())
			return 1
		}
	}
	profile, err := ioutil.ReadFile(c.fileloc)
	if err != nil {
		c.UI.Error(fmt.Sprintf("Could not read file at %s \nError: %s", c.fileloc, err.Error()))
		return 1
	}
	creds.AppleSecrets = profile
	creds.Identifier = c.identifier
	creds.AppleSecretsPassword = c.profPw

	if err = uploadCredential(ctx, c.config.Client, c.UI, creds); err != nil {
		if _, ok := err.(*commandhelper.DontOverwrite); ok {
			return 0
		}
		c.UI.Error("Could not add Apple Dev Profile credential to admin")
		commandhelper.UIErrFromGrpc(err, c.UI, err.Error())
		return 1
	}
	c.UI.Info("Successfully added an apple developer profile to the account " + c.account)
	return 0
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return help
}

const synopsis = "Add an apple developer profile for authorization to sign certificates in ocelot builds"
const help = `
Usage: ocelot creds apple add -acct my_kewl_acct -zip=/Users/jessishank/jessdev.developerprofile

`

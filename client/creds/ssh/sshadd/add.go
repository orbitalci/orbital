package sshadd

import (
	"context"
	"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"github.com/level11consulting/ocelot/client/commandhelper"
	models "github.com/level11consulting/ocelot/models/pb"
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
	sshKeyFile string
	acctName   string
	identifier string
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
	c.flags.StringVar(&c.sshKeyFile, "sshfile-loc", "ERROR", "location of ssh private key to upload")
	c.flags.StringVar(&c.acctName, "acctname", "ERROR", "account name matching with sshfile-loc")
	c.flags.StringVar(&c.identifier, "identifier", "ERROR", "unique identifier for this ssh key")
}

// uploadCredential will check if credential already exists. if it does, it will ask if the user wishes to overwrite. if the user responds YES, the credential will be updated.
// if it does not exist, will be inserted as normal.
func uploadCredential(ctx context.Context, client models.GuideOcelotClient, UI cli.Ui, cred *models.SSHKeyWrapper) error {
	exists, err := client.SSHCredExists(ctx, cred)
	if err != nil {
		return err
	}

	if exists.Exists {
		update, err := UI.Ask(fmt.Sprintf("Entry with Account Name %s, SSH Type %s, and Idnetifier %s already exists. Do you want to overwrite? "+
			"Only a YES will continue with update, otherwise the client will exit. ", cred.AcctName, strings.ToLower(cred.SubType.String()), cred.Identifier))
		if err != nil {
			return err
		}
		if update != "YES" {
			UI.Info("Did not recieve a YES at the prompt, will not overwrite. Exiting.")
			return &commandhelper.DontOverwrite{}
		}
		_, err = client.UpdateSSHCreds(ctx, cred)
		if err != nil {
			return err
		}
		UI.Info("Succesfully updated SSH Credential.")
		return nil
	}
	_, err = client.SetSSHCreds(ctx, cred)
	if err != nil {
		UI.Info("Successfully added SSH Credential.")
	}
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
	if c.sshKeyFile == "ERROR" || c.acctName == "ERROR" || c.identifier == "ERROR" {
		c.UI.Error("--sshfile-loc, --acctname, and --identifier are required")
		return 1
	}
	cred := &models.SSHKeyWrapper{
		AcctName:   c.acctName,
		Identifier: c.identifier,
		SubType:    models.SubCredType_SSHKEY,
	}
	sshKey, err := ioutil.ReadFile(c.sshKeyFile)
	if err != nil {
		c.UI.Error(fmt.Sprintf("\tCould not read file at %s \nError: %s", c.sshKeyFile, err.Error()))
	}
	cred.PrivateKey = sshKey
	err = uploadCredential(ctx, c.GetClient(), c.UI, cred)
	if err != nil {
		c.UI.Error("An error occurred adding SSH Key. Error: " + err.Error())
		return 1
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
Usage: ocelot creds ssh add --identifier JESSI_SSH_KEY --acctname level11consulting --sshfile-loc /Users/jesseshank/.ssh/id_rsa
  Will add an ssh key and attach it to an account. In the useage above, the ssh key will be accessible within the script as ~/.ssh/JESSI_SSH_KEY to use as an identity file. 
`

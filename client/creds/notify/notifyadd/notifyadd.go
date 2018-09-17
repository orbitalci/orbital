package notifyadd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/shankj3/ocelot/client/commandhelper"
	models "github.com/shankj3/ocelot/models/pb"
)

func New(ui cli.Ui) *cmd {
	c := &cmd{UI: ui, config: commandhelper.Config}
	c.init()
	return c
}

type cmd struct {
	UI            cli.Ui
	flags         *flag.FlagSet
	notifyUrl     string
	acctName      string
	identifier    string
	stringSubType string
	baseUrl       string
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
	c.flags.StringVar(&c.notifyUrl, "url", "ERROR", "url associated with this type of notification")
	// right now only slack is supported
	// todo: change this when more support!
	c.flags.StringVar(&c.stringSubType, "type", "slack", "type of notify cred, currently only SLACK")
	c.flags.StringVar(&c.acctName, "acctname", "ERROR", "account name to associate notify cred witth")
	c.flags.StringVar(&c.identifier, "identifier", "ERROR", "unique identifier for this notify cred on this account")
	c.flags.StringVar(&c.baseUrl, "detail-url", "", "[optional] base url for ocelot web ui")
}

// uploadCredential will check if credential already exists. if it does, it will ask if the user wishes to overwrite. if the user responds YES, the credential will be updated.
// if it does not exist, will be inserted as normal.
func uploadCredential(ctx context.Context, client models.GuideOcelotClient, UI cli.Ui, cred *models.NotifyCreds) error {
	exists, err := client.NotifyCredExists(ctx, cred)
	if err != nil {
		return err
	}

	if exists.Exists {
		update, err := UI.Ask(fmt.Sprintf("Entry with Account Name %s,  Type %s, and Idnetifier %s already exists. Do you want to overwrite? "+
			"Only a YES will continue with update, otherwise the client will exit. ", cred.AcctName, strings.ToLower(cred.SubType.String()), cred.Identifier))
		if err != nil {
			return err
		}
		if update != "YES" {
			UI.Info("Did not recieve a YES at the prompt, will not overwrite. Exiting.")
			return &commandhelper.DontOverwrite{}
		}
		_, err = client.UpdateNotifyCreds(ctx, cred)
		if err != nil {
			return err
		}
		UI.Info("Succesfully updated Notify Credential.")
		return nil
	}
	_, err = client.SetNotifyCreds(ctx, cred)
	if err != nil {
		UI.Info("Successfully added Notify Credential.")
	}
	return err
}

func getSubTypeFromString(str string) (models.SubCredType, error) {
	subcred, ok := models.SubCredType_value[strings.ToUpper(str)]
	if !ok {
		return 0, errors.New("could not parse type string, needs to be in 'SLACK|slack'")
	}
	return models.SubCredType(subcred), nil
}

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	c.UI.Warn("Currently --type will default to slack as that is the only notify credential. This will not always be the case.")
	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}
	if c.notifyUrl == "ERROR" || c.acctName == "ERROR" || c.identifier == "ERROR" {
		c.UI.Error("--url, --acctname, and --identifier are required")
		return 1
	}
	sct, err := getSubTypeFromString(c.stringSubType)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	cred := &models.NotifyCreds{
		AcctName:      c.acctName,
		Identifier:    c.identifier,
		SubType:       sct,
		ClientSecret:  c.notifyUrl,
		DetailUrlBase: c.baseUrl,
	}
	err = uploadCredential(ctx, c.GetClient(), c.UI, cred)
	if err != nil {
		c.UI.Error("An error occurred adding Notify Credential. Error: " + err.Error())
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

const synopsis = "Add notify credentials"
const help = `
Usage: ocelot creds notify add --identifier L11_SLACK --acctname level11consulting --url https://hooks.slack.com/services/T0DFsdSBA/345PPRP9C/5hUe12345v6BrxfSJt
  Currently only slack

  Example: using this notify credential in your ocelot.yml would look like this:
	notify:
      slack:
        channel: "@jessishank"
        identifier: "L11_SLACK"
        on:
        - "PASS"
        - "FAIL"
`

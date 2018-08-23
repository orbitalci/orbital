package delete

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/shankj3/ocelot/client/commandhelper"
	models "github.com/shankj3/ocelot/models/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)


func New(ui cli.Ui, credType models.CredType) *cmd {
	c := &cmd{UI: ui, config: commandhelper.Config, credType: credType}
	c.init()
	return c
}

type cmd struct {
	UI         cli.Ui
	flags      *flag.FlagSet
	st         string
	account    string
	identifier string
	credType   models.CredType
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
	c.flags.StringVar(&c.st, "subtype", "ERROR", "If more than one subtype for this cred, specify subtype")
	c.flags.StringVar(&c.account, "acct", "ERROR", "Account name to file the xcode profile under.")
	c.flags.StringVar(&c.identifier, "identifier", "ERROR","unique identifier for this ssh key")
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

	if c.identifier == "ERROR" {
		c.UI.Error("-identifier required")
		return 1
	}
	availableTypesString := strings.Join(c.credType.SubtypesString(), "|")
	var subType models.SubCredType
	if c.st == "ERROR" {
		if len(c.credType.Subtypes()) == 1 {
			subType = c.credType.Subtypes()[0]
		} else {
			c.UI.Error(fmt.Sprintf("-subtype not provided and %s has more than one subtype, please select %s", c.credType.String(), availableTypesString))
			return 1
		}
	} else {
		stt, ok := models.SubCredType_value[strings.ToUpper(c.st)]
		if !ok {
			c.UI.Error(fmt.Sprintf("Not found in the subtype map, please check spelling. \nYours: %s\nAvailable: %s", c.st, availableTypesString))
			return 1
		}
		subType = models.SubCredType(stt)
		if !models.Contains(subType, c.credType.Subtypes()) {
			c.UI.Error(fmt.Sprintf("Not the correct subtype for this credential (%s). Please pick one of %s", c.credType.String(), availableTypesString))
			return 1
		}
	}
	yes, err := c.UI.Ask("Are you sure that you want to delete this credential? This action is irreversible! Type YES if you mean it.")
	if err != nil {
		c.UI.Error("Error occured, exiting.. \nError: " + err.Error())
		return 1
	}
	if yes != "YES" {
		c.UI.Info("YES not entered, not deleting credential... ")
		return 0
	}
	return c.DeleteACredential(ctx, subType)
}

func (c *cmd) DeleteACredential(ctx context.Context, subType models.SubCredType) int {
	var err error
	switch c.credType {
	case models.CredType_GENERIC:
		_, err = c.config.Client.DeleteGenericCreds(ctx, &models.GenericCreds{AcctName: c.account, Identifier: c.identifier, SubType:subType})
	case models.CredType_NOTIFIER:
		_, err = c.config.Client.DeleteNotifyCreds(ctx, &models.NotifyCreds{AcctName: c.account, Identifier: c.identifier, SubType: subType})
	case models.CredType_SSH:
		_, err = c.config.Client.DeleteSSHCreds(ctx, &models.SSHKeyWrapper{AcctName: c.account, Identifier: c.identifier, SubType: subType})
	case models.CredType_REPO:
		_, err = c.config.Client.DeleteRepoCreds(ctx, &models.RepoCreds{AcctName: c.account, Identifier: c.identifier, SubType: subType})
	case models.CredType_K8S:
		_, err = c.config.Client.DeleteK8SCreds(ctx, &models.K8SCreds{AcctName: c.account, Identifier: c.identifier, SubType: subType})
	default:
		c.UI.Error("Not currently supported for cred deletion.")
		return 1
	}
	if err == nil {
		c.UI.Info(fmt.Sprintf("Successfully deleted %s credential under account %s and identifier %s", subType.String(), c.account, c.identifier))
		return 0
	}
	statErr, ok := status.FromError(err)
	if !ok {
		c.UI.Error("Unexpected error deleting credential! Error: " + err.Error())
		return 1
	}
	if statErr.Code() == codes.NotFound {
		c.UI.Error("Credential not found")
		return 1
	}
	c.UI.Error(fmt.Sprintf("An error (%s) has occured: %s", subType.String(), statErr.Message()))
	return 1
}

func (c *cmd) Synopsis() string {
	return fmt.Sprintf(synopsis, c.credType.String())
}

func (c *cmd) Help() string {
	return help
}

const synopsis = "Delete a credential of %s type"
const help = `
Usage: ocelot creds ssh delete -account level11consulting -identifier woops -subtype SSH
       ocelot creds repo delete -account shankj3 -identifier docker_credss -subtype docker
`

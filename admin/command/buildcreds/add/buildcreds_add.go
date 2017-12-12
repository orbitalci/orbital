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
}

func (c *cmd) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.StringVar(&c.fileloc, "credfile-loc", "",
		"Location of yaml file containing creds to upload")
}

/*
type Credentials struct {
	ClientId     string `protobuf:"bytes,1,opt,name=clientId" json:"clientId,omitempty"`
	ClientSecret string `protobuf:"bytes,2,opt,name=clientSecret" json:"clientSecret,omitempty"`
	TokenURL     string `protobuf:"bytes,3,opt,name=tokenURL" json:"tokenURL,omitempty"`
	AcctName     string `protobuf:"bytes,4,opt,name=acctName" json:"acctName,omitempty"`
	Type         string `protobuf:"bytes,5,opt,name=type" json:"type,omitempty"`
}
 */

func (c *cmd) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	client, err := admin.GetClient("localhost:10000")
	ctx := context.Background()
	if err != nil {
		panic(err)
	}

	if c.fileloc != "" {
		credWrap := &models.CredWrapper{}
		dese := deserialize.New()
		confFile, err := ioutil.ReadFile(c.fileloc)
		if err != nil {
			fmt.Println("Could not read file at ", c.fileloc)
			fmt.Println("Error: ", err)
			return 1
		}
		if err = dese.YAMLToProto(confFile, credWrap); err != nil {
			fmt.Println("Could not process file, please check documentation")
			fmt.Println("Error", err)
			return 1
		}
		var errOccured bool
		if len(credWrap.Credentials) == 0 {
			fmt.Println("Did not read any credentials! Is your yaml formatted correctly?")
			return 1
		}
		for _, configVal := range credWrap.Credentials {
			_, err = client.SetCreds(ctx, configVal)
			if err != nil {
				fmt.Println("Could not add credentials for account: ", configVal.AcctName)
				fmt.Println("Error: ", err)
				errOccured = true
			} else {
				fmt.Println("Added credentials for account: ", configVal.AcctName)
			}
		}
		if errOccured {
			return 1
		}

	} else {
		creds := &models.Credentials{}
		creds.ClientId, err = c.UI.Ask("Client ID: ")
		creds.Type, err = c.UI.Ask("Type: ")
		creds.AcctName, err = c.UI.Ask("Account Name: ")
		creds.TokenURL, err = c.UI.Ask("OAuth token URL for repository: ")
		creds.ClientSecret, err = c.UI.AskSecret("Secret for OAuth: ")
		if err != nil {
			fmt.Println("Error recieving input: \n ", err)
			return 1
		}
		if _, err = client.SetCreds(ctx, creds); err != nil {
			fmt.Println("Could not add credentials for account: ", creds.AcctName)
			fmt.Println("Error: ", err)
			return 1
		}
		fmt.Println("Successfully added credentials for account: ", creds.AcctName)
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

package ocevault

import (
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/vault/api"
	"os"
)

// VaultCIPath is the base path for vault. Will be formatted to include the user or group when
// setting or retrieving credentials.
var VaultCIPath = "secrets/ci/%s"
var Token = "04eeeacd-5846-36cb-f885-cf9700d84f45"

// NewEnvAuthedClient will set the Client token based on the environment variable `$VAULT_TOKEN`.
// Will return error if it is not set.
func NewEnvAuthClient() (*api.Client, error) {
	var token string
	if token = os.Getenv("VAULT_TOKEN"); token == "" {
		return &api.Client{}, errors.New("$VAULT_TOKEN not set")
	}
	return NewAuthedClient(token)
}

// NewAuthedClient will return a client with default configurations and the Token attached to it.
// Vault URL configured through VAULT_ADDR environment variable.
func NewAuthedClient(token string) (cli *api.Client, err error) {
	config := api.DefaultConfig()
	cli, err = api.NewClient(config)
	if err != nil {
		return
	}
	cli.SetToken(token)
	return
}

// AddUserAuthData will add the values of the data map to the path of the CI user creds
// CI vault path set off of base path VaultCIPath
// Expects that the Client is already initialized properly, will error out if not.
func AddUserAuthData(cli *api.Client, user string, data map[string]interface{}) (*api.Secret, error){
	return cli.Logical().Write(fmt.Sprintf(VaultCIPath, user), data)
}

// GetSecretData will return the Data attribute of the secret you get at the path of the CI user creds, ie all the
// key-value fields that were set on it. Expects that the Client is already initialized properly (has the
// appropriate token and URL)
func GetUserAuthData(cli *api.Client, user string) (map[string]interface{}, error) {
	secret, err := cli.Logical().Read(fmt.Sprintf(VaultCIPath, user))
	if err != nil {
		return make(map[string]interface{}), err
	}
	return secret.Data, nil
}

// Test function
func Do() {
	cli, err := NewAuthedClient(Token)
	if err != nil {
		panic("boooOOoooooOOoooOOOoo")
	}
	v, _ := cli.Logical().Read("secret/booboo")
	spew.Dump(v.Data)
}
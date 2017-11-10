package ocevault

import (
	"errors"
	"fmt"
	//"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/vault/api"
	"os"
)

// VaultCIPath is the base path for vault. Will be formatted to include the user or group when
// setting or retrieving credentials.
var VaultCIPath = "secrets/ci/creds/%s"
var Token = "e57369ad-9419-cc03-9354-fc227b06f795"

// Some blog said that changing any *api.Client functions to take in a n interface instead
// will make testing easier. I agree, just have to figure out how to do this properly without
// wasting memory

//type ApiClient interface {
//	Logical() *ApiLo
//
//}
//
//type ApiLogical interface {
//	Read(path string) (*api.Secret, error)
//	Write(path string, data map[string]interface{}) (*api.Secret, error)
//}

type Ocevault struct {
	Client	*api.Client
	Config	*api.Config
}


// NewEnvAuthedClient will set the Client token based on the environment variable `$VAULT_TOKEN`.
// Will return error if it is not set. Returns configured ocevault struct
func NewEnvAuthClient() (*Ocevault, error) {
	var token string
	if token = os.Getenv("VAULT_TOKEN"); token == "" {
		return &Ocevault{}, errors.New("$VAULT_TOKEN not set")
	}
	return NewAuthedClient(token)
}

// NewAuthedClient will return a client with default configurations and the Token attached to it.
// Vault URL configured through VAULT_ADDR environment variable.
func NewAuthedClient(token string) (oce *Ocevault, err error) {
	oce = &Ocevault{}
	oce.Config = api.DefaultConfig()
	oce.Client, err = api.NewClient(oce.Config)
	if err != nil {
		return
	}
	oce.Client.SetToken(token)
	return
}

// AddUserAuthData will add the values of the data map to the path of the CI user creds
// CI vault path set off of base path VaultCIPath
func (oce *Ocevault) AddUserAuthData(user string, data map[string]interface{}) (*api.Secret, error){
	return oce.Client.Logical().Write(fmt.Sprintf(VaultCIPath, user), data)
}

// GetSecretData will return the Data attribute of the secret you get at the path of the CI user creds, ie all the
// key-value fields that were set on it
func (oce *Ocevault) GetUserAuthData(user string) (map[string]interface{}, error){
	secret, err := oce.Client.Logical().Read(fmt.Sprintf(VaultCIPath, user))
	if err != nil {
		return nil, err
	}
	return secret.Data, nil
}

func (oce *Ocevault) CreateThrowawayToken() (*api.Secret, error) {
	return nil, nil
}

//
//// Test function
//func Do() {
//	cli, err := NewAuthedClient(Token)
//	if err != nil {
//		panic("boooOOoooooOOoooOOOoo")
//	}
//	v, _ := cli.Logical().Read("secret/booboo")
//	spew.Dump(v.Data)
//}
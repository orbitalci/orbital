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
var VaultCIPath = "secret/ci/creds/%s"
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
	// this action is idempotent, and since we *need* this policy for generating tokens, might as well?
	// i guess?
	oce.CreateOcevaultPolicy()
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
	path := fmt.Sprintf(VaultCIPath, user)
	secret, err := oce.Client.Logical().Read(path)
	if err != nil {
		return nil, err
	}
	if secret == nil {
		return nil, errors.New(fmt.Sprintf("User data not found, path searched: %s", path))
	}
	return secret.Data, nil
}

// CreateToken creates an Auth token using the oce.Client's creds. Look at api.TokenCreateRequest docs
// for how to configure the token. Will return any errors from the create request.
func (oce *Ocevault) CreateToken(request *api.TokenCreateRequest) (token string, err error) {
	secret, err := oce.Client.Auth().Token().Create(request)
	if err != nil {
		return
	}
	token = secret.Auth.ClientToken
	return
}

// CreateThrowawayToken creates a single use token w/ same privileges as client.
// *single use* really means enough uses to initialize the client and make one call to actually
// get data
// todo: add ocevault policy for reading the secrets/ci/user path
func (oce *Ocevault) CreateThrowawayToken() (token string, err error) {
	tokenReq := &api.TokenCreateRequest{
		//Policies: 		[]string{"ocevault"}, // todo: figure out why this doesn't work...
		TTL:            "30m",
		NumUses:		3,
	}
	//oce.Client.Auth().Token().Create(&api.})
	return oce.CreateToken(tokenReq)
}

// CreateOcevaultPolicy creates a policy for r/w ops on only the path that credentials are on, which is `secret/ci/creds`.
// Tokens that are one-off and passed to the workers for building will get this access.
func (oce *Ocevault) CreateOcevaultPolicy() error {
	err := oce.Client.Sys().PutPolicy("ocevault", "path \"secret/ci/creds\" {\n capabilities = [\"read\", \"list\"]\n}")
	if err != nil {
		return err
	}
	return nil
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
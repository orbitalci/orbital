package credentials

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	ocevault "github.com/shankj3/go-til/vault"
)

// getVaultAddr will set the Vault address in this order:
// Passing in Vault through command line options takes priority
// If not passed in, the VAULT_ADDR environment variable is next
// If not defined, assume http://localhost:8200
//func getVaultAddr() error {
//}

// getToken will check for a vault token first in the environment variable VAULT_TOKEN. If not found at the env var,
// either the path given or the default path of /etc/vaulted/token will be searched to see if it exists. If it exists,
// its contents will be read and returned as the vault token. for kubernetes support
func getVaultToken(path string) (string, error) {
	defaultPath := "/etc/vaulted/token"
	if path == "" {
		path = defaultPath
	}
	var token string
	if token = os.Getenv("VAULT_TOKEN"); token != "" {
		return token, nil
	} else {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return "", errors.New("$VAULT_TOKEN not set and no token found at filepath " + path)
		}
		tokenB, err := ioutil.ReadFile(path)
		if err != nil {
			return "", errors.WithMessage(err, fmt.Sprintf("File exists at %s but could not be read", path))
		}
		return strings.TrimSpace(string(tokenB)), nil
	}

}

// GetVault returns the local vault client handler
func (rc *RemoteConfig) GetVault() ocevault.Vaulty {
	return rc.Vault
}

// SetVault sets the local vault client handler
func (rc *RemoteConfig) SetVault(vault ocevault.Vaulty) {
	rc.Vault = vault
}

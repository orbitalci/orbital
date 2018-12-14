package config

import (
	"github.com/pkg/errors"
)

// AddSSHKey adds repo ssh private key to vault at the usual vault path + /ssh
func (rc *RemoteConfig) AddSSHKey(path string, sshKeyFile []byte) (err error) {
	if rc.Vault != nil {
		secret := buildSecretPayload(string(sshKeyFile))
		if _, err = rc.Vault.AddUserAuthData(path+"/ssh", secret); err != nil {
			return
		}
	} else {
		err = errors.New("no connection to vault, unable to add SSH Key")
	}
	return
}

// CheckSSHKeyExists returns a boolean indicating whether or not an ssh key has been uploaded
func (rc *RemoteConfig) CheckSSHKeyExists(path string) error {
	var err error

	if rc.Vault != nil {
		_, err := rc.Vault.GetUserAuthData(path + "/ssh")
		if err != nil {
			return err
		}
	} else {
		err = errors.New("no connection to vault, unable to add SSH Key")
	}

	return err
}

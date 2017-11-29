package util

import (
	"fmt"
	"github.com/shankj3/ocelot/admin/models"
	"github.com/shankj3/ocelot/util/consulet"
	"github.com/shankj3/ocelot/util/ocelog"
	"github.com/shankj3/ocelot/util/ocevault"
	"strings"
	"github.com/pkg/errors"
)

var ConfigPath = "creds"

//GetInstance returns a new instance of ConfigConsult. If consulHot and consulPort are empty,
//this will talk to consul using reasonable defaults (localhost:8500)
//if token is an empty string, vault will be initialized with $VAULT_TOKEN
func GetInstance(consulHost string, consulPort int, token string) (*RemoteConfig, error) {
	remoteConfig := &RemoteConfig{}

	//intialize consul
	if consulHost == "" && consulPort == 0 {
		remoteConfig.Consul = consulet.Default()
	} else {
		remoteConfig.Consul = consulet.New(consulHost, consulPort)
	}

	//initialize vault
	if token == "" {
		vaultClient, err := ocevault.NewEnvAuthClient()
		if err != nil {
			return nil, err
		}
		remoteConfig.Vault = vaultClient
	} else {
		vaultClient, err := ocevault.NewAuthedClient(token)
		if err != nil {
			return nil, err
		}
		remoteConfig.Vault = vaultClient
	}

	return remoteConfig, nil
}

//RemoteConfig is an abstraction for retrieving/setting creds for ocelot
//currently uses consul + vault
type RemoteConfig struct {
	Consul *consulet.Consulet
	Vault  *ocevault.Ocevault
}

//GetCredAt will return list of credentials stored at specified path.
//if hideSecret is set to false, will return password in cleartext
//key of map is CONFIG_TYPE/ACCTNAME. Ex: bitbucket/mariannefeng
//if an error occurs while reading from vault, the most recent error will be returned from the response
func (remoteConfig *RemoteConfig) GetCredAt(path string, hideSecret bool) (map[string]*models.AdminConfig, error) {
	creds := map[string]*models.AdminConfig{}
	var err error

	if remoteConfig.Consul.Connected {
		configs, err := remoteConfig.Consul.GetKeyValues(path)
		if err != nil {
			return creds, err
		}

		for _, v := range configs {
			pathKeys := strings.Split(strings.TrimLeft(v.Key, "/" + ConfigPath), "/")

			//cred type | acct name gives us a unique id to track by in the map
			credType := pathKeys[0]
			acctName := pathKeys[1]
			infoType := pathKeys[2]

			mapKey := credType + "/" + acctName
			foundConfig, ok := creds[mapKey]
			if !ok {
				foundConfig = &models.AdminConfig{
					AcctName: acctName,
					Type:     credType,
				}

				if hideSecret {
					foundConfig.ClientSecret = "*********"
				} else {
					passcode, passErr := remoteConfig.GetPassword(ConfigPath + "/" + credType + "/" + acctName)
					if passErr != nil {
						ocelog.IncludeErrField(err).Error()
						foundConfig.ClientSecret = "ERROR: COULD NOT RETRIEVE PASSWORD FROM VAULT"
						err = passErr
					} else {
						foundConfig.ClientSecret = passcode
					}
				}

				creds[mapKey] = foundConfig
			}

			switch infoType {
			case "clientid":
				foundConfig.ClientId = string(v.Value[:])
			case "tokenurl":
				foundConfig.TokenURL = string(v.Value[:])
			}
		}
	} else {
		return creds, errors.New("not connected to consul, unable to retrieve credentials")
	}

	return creds, err
}

//GetPassword will return to you the vault password at specified path
func (remoteConfig *RemoteConfig) GetPassword(path string) (string, error) {
	path = "secret/ci/" + path
	authData, err := remoteConfig.Vault.GetUserAuthData(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", authData["clientsecret"]), nil
}

//AddCreds adds your adminconfig creds into both consul + vault
func (remoteConfig *RemoteConfig) AddCreds(path string, adminConfig *models.AdminConfig) error {
	if remoteConfig.Consul.Connected {
		err := remoteConfig.Consul.AddKeyValue(path+"/clientid", []byte(adminConfig.ClientId))
		if err != nil {
			return err
		}
		err = remoteConfig.Consul.AddKeyValue(path+"/tokenurl", []byte(adminConfig.TokenURL))
		if err != nil {
			return err
		}
		if remoteConfig.Vault != nil {
			secret := make(map[string]interface{})
			secret["clientsecret"] = adminConfig.ClientSecret
			_, err := remoteConfig.Vault.AddUserAuthData("secret/ci/" + path, secret)
			if err != nil {
				return err
			}
		}
	} else {
		return errors.New("not connected to consul, unable to add credentials")
	}

	return nil
}

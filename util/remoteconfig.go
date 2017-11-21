package util

import (
	"fmt"
	"github.com/shankj3/ocelot/admin/models"
	"github.com/shankj3/ocelot/util/consulet"
	"github.com/shankj3/ocelot/util/ocelog"
	"github.com/shankj3/ocelot/util/ocevault"
	"strings"
)

var ConfigPath = "creds"

//GetInstance returns a new instance of ConfigConsult
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
func (remoteConfig *RemoteConfig) GetCredAt(path string, hideSecret bool) map[string]*models.AdminConfig {
	creds := map[string]*models.AdminConfig{}
	for _, v := range remoteConfig.Consul.GetKeyValues(path) {
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
				passcode, err := remoteConfig.GetPassword(ConfigPath + "/" + credType + "/" + acctName)
				if err != nil {
					ocelog.IncludeErrField(err).Error()
					foundConfig.ClientSecret = "ERROR: COULD NOT RETRIEVE PASSWORD FROM VAULT"
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
	return creds
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
		remoteConfig.Consul.AddKeyValue(path+"/clientid", []byte(adminConfig.ClientId))
		remoteConfig.Consul.AddKeyValue(path+"/tokenurl", []byte(adminConfig.TokenURL))
		if remoteConfig.Vault != nil {
			secret := make(map[string]interface{})
			secret["clientsecret"] = adminConfig.ClientSecret
			_, err := remoteConfig.Vault.AddUserAuthData("secret/ci/" + path, secret)
			if err != nil {
				return err
			}
		}
	} else {
		ocelog.Log().Error("NOT CONNECTED TO CONSUL")
	}
	return nil
}

package cred

import (
	"bitbucket.org/level11consulting/go-til/consul"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	ocevault "bitbucket.org/level11consulting/go-til/vault"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

var (
	ConfigPath = "creds"
 	VCSPath = ConfigPath + "/vcs"
 	RepoPath = ConfigPath + "/repo"
)

// OcyCredType is the the of credential that we will be storing, ie binary repo or vcs
type OcyCredType int

const (
	Vcs OcyCredType = iota
	Repo
)

var OcyCredMap = map[string]OcyCredType{
"vcs": Vcs,
"repo": Repo,
}


func BuildCredPath(credType string, AcctName string, ocyCredType OcyCredType) string {
	var pattern string
	switch ocyCredType {
	case Vcs: pattern = "%s/vcs/%s/%s"
	case Repo: pattern = "%s/repo/%s/%s"
	default: panic("only repo or vcs")
	}
	return fmt.Sprintf(pattern, ConfigPath, AcctName, credType)
}


//GetInstance returns a new instance of ConfigConsult. If consulHot and consulPort are empty,
//this will talk to consul using reasonable defaults (localhost:8500)
//if token is an empty string, vault will be initialized with $VAULT_TOKEN
func GetInstance(consulHost string, consulPort int, token string) (*RemoteConfig, error) {
	remoteConfig := &RemoteConfig{}

	//intialize consul
	if consulHost == "" && consulPort == 0 {
		consulet, err := consul.Default()
		if err != nil {
			return nil, err
		}
		remoteConfig.Consul = consulet
	} else {
		consulet, err := consul.New(consulHost, consulPort)
		remoteConfig.Consul = consulet
		if err != nil {
			return nil, err
		}
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
	Consul *consul.Consulet
	Vault  *ocevault.Vaulty
}

func splitConsulCredPath(path string) (string,string,string,string) {
	pathKeys := strings.Split(strings.TrimLeft(path, "/"+ConfigPath), "/")
	return pathKeys[0], pathKeys[1], pathKeys[2], pathKeys[3]
}

// GetRepoCredAt will return repo credentials for repo integrations. its waaay to similar to getCredAt, so this is
// unacceptable. i can't figure out how to successfully use map[string]interface{}
func (remoteConfig *RemoteConfig) GetRepoCredAt(path string, hideSecret bool) (creds map[string]*models.RepoCreds, err error) {
	if remoteConfig.Consul.Connected {
		configs, err := remoteConfig.Consul.GetKeyValues(path)
		if err != nil {
			return
		}
		for _, v := range configs {
			_, acctName, credType, infoType := splitConsulCredPath(v.Key)
			mapKey := credType + "/" + acctName
			foundConfig, ok := creds[mapKey]
			if !ok {
				foundConfig := &models.RepoCreds{
					AcctName: acctName,
					Type: infoType,
				}
				if hideSecret {
					foundConfig.Password = "*********"
				} else {
					passcode, passErr := remoteConfig.GetPassword(BuildCredPath(credType, acctName, Vcs))
					//passcode, passErr := remoteConfig.GetPassword(ConfigPath + "/" + credType + "/" + acctName)
					if passErr != nil {
						ocelog.IncludeErrField(err).Error()
						foundConfig.Password = "ERROR: COULD NOT RETRIEVE PASSWORD FROM VAULT"
						err = passErr
					} else {
						foundConfig.Password = passcode
					}
				}
			}
			switch infoType {
			case "repourl":
				foundConfig.RepoUrl = string(v.Value[:])
			case "username":
				foundConfig.Username = string(v.Value[:])
			}
		}
	} else {
		return creds, errors.New("not connected to consul, unable to retrieve credentials")
	}
	return creds, err
}


//GetCredAt will return map of credentials stored at specified path.
//if hideSecret is set to false, will return password in cleartext
//key of map is cred_type/acct_name. Ex: bitbucket/mariannefeng
//if an error occurs while reading from vault, the most recent error will be returned from the response
func (remoteConfig *RemoteConfig) GetCredAt(path string, hideSecret bool) (creds map[string]*models.Credentials, err error) {
	if remoteConfig.Consul.Connected {
		configs, err := remoteConfig.Consul.GetKeyValues(path)
		if err != nil {
			return creds, err
		}
		for _, v := range configs {
			_, acctName, credType, infoType := splitConsulCredPath(v.Key)
			mapKey := credType + "/" + acctName
			foundConfig, ok := creds[mapKey]
			if !ok {
				foundConfig = &models.Credentials{
					AcctName: acctName,
					Type:     credType,
				}
				if hideSecret {
					foundConfig.ClientSecret = "*********"
				} else {
					passcode, passErr := remoteConfig.GetPassword(BuildCredPath(credType, acctName, Vcs))
					//passcode, passErr := remoteConfig.GetPassword(ConfigPath + "/" + credType + "/" + acctName)
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
	authData, err := remoteConfig.Vault.GetUserAuthData(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", authData["clientsecret"]), nil
}

//AddCreds adds your adminconfig creds into both consul + vault
func (remoteConfig *RemoteConfig) AddCreds(path string, adminConfig *models.Credentials) error {
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
			_, err := remoteConfig.Vault.AddUserAuthData(path, secret)
			if err != nil {
				return err
			}
		}
	} else {
		return errors.New("not connected to consul, unable to add credentials")
	}

	return nil
}

// AddRepoCreds adds repo integration creds to consul + vault
func (remoteConfig *RemoteConfig) AddRepoCreds(path string, repoCred *models.RepoCreds) (err error) {
	if remoteConfig.Consul.Connected {
		if err = remoteConfig.Consul.AddKeyValue(path + "/username", []byte(repoCred.Username)); err != nil {
			return
		}
		if err = remoteConfig.Consul.AddKeyValue(path + "/repourl", []byte(repoCred.RepoUrl)); err != nil {
			return
		}
		if remoteConfig.Vault != nil {
			secret := make(map[string]interface{})
			secret["clientsecret"] = repoCred.Password
			if _, err = remoteConfig.Vault.AddUserAuthData(path, secret); err != nil {
				return
			}
		}
	} else {
		err = errors.New("not connected to consul, unable to add credentials for artifact repository")
	}
	return
}

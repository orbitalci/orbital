package cred

import (
	"bitbucket.org/level11consulting/go-til/consul"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	ocevault "bitbucket.org/level11consulting/go-til/vault"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"fmt"
	"github.com/pkg/errors"
)

var (
	ConfigPath = "creds"
 	VCSPath = ConfigPath + "/vcs"
 	RepoPath = ConfigPath + "/repo"
)

//GetInstance returns a new instance of ConfigConsult. If consulHot and consulPort are empty,
//this will talk to consul using reasonable defaults (localhost:8500)
//if token is an empty string, vault will be initialized with $VAULT_TOKEN
func GetInstance(consulHost string, consulPort int, token string) (CVRemoteConfig, error) {
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

func New() (CVRemoteConfig, error) {
	return GetInstance("localhost", 8500, "")
}


//RemoteConfig is an abstraction for retrieving/setting creds for ocelot
//currently uses consul + vault
type CVRemoteConfig interface {
	GetConsul()	*consul.Consulet
	SetConsul(consul *consul.Consulet)
	GetVault() ocevault.Vaulty
	SetVault(vault ocevault.Vaulty)
	CheckExists(path string) error
	GetCredAt(path string, hideSecret bool, ocyType OcyCredType) (map[string]RemoteConfigCred, error)
	GetPassword(path string) (string, error)
	AddCreds(path string, anyCred RemoteConfigCred) (err error)
}

type RemoteConfig struct {
	Consul *consul.Consulet
	Vault  ocevault.Vaulty
}

func (rc *RemoteConfig) GetConsul() *consul.Consulet {
	return rc.Consul
}

func (rc *RemoteConfig) SetConsul(consul *consul.Consulet) {
	rc.Consul = consul
}

func (rc *RemoteConfig)  GetVault() ocevault.Vaulty {
	return rc.Vault
}

func (rc *RemoteConfig) SetVault(vault ocevault.Vaulty) {
	rc.Vault = vault
}


// instantiateCredObject is what we will have to add too when we add new credential integrations
// (ie slack, w/e)
// todo: find out a way to use either this method or GetCredAt to remove the cred package's dependency on models
func instantiateCredObject(ocyType OcyCredType) RemoteConfigCred {
	switch ocyType {
	case Vcs:
		return &models.VCSCreds{}
	case Repo:
		return &models.RepoCreds{}
	default:
		panic("ahh!")
	}
}

// GetCred at will return a map w/ key <cred_type>/<acct_name> to credentials. depending on the OcyCredType,
//   the appropriate credential struct will be instantiated and filled with data from consul and vault.
//   currently supports map[string]*models.VCSCreds and map[string]*models.RepoCreds
//   You must cast the resulting values to their appropriate objects after the map is generated if you need to access more than
//   the methods on the cred.RemoteConfigCred interface
//   Example:
//      creds, err := g.RemoteConfig.GetCredAt(cred.VCSPath, true, cred.Vcs)
//      vcsCreds := creds.(*models.VCSCreds)
func (rc *RemoteConfig) GetCredAt(path string, hideSecret bool, ocyType OcyCredType) (map[string]RemoteConfigCred, error) {
	creds := map[string]RemoteConfigCred{}
	var err error
	if rc.Consul.Connected {
		configs, err := rc.Consul.GetKeyValues(path)
		if err != nil {
			return creds, err
		}
		for _, v := range configs {
			_, acctName, credType, infoType := splitConsulCredPath(v.Key)
			mapKey := credType + "/" + acctName
			foundConfig, ok := creds[mapKey]
			if !ok {
				foundConfig = instantiateCredObject(ocyType)
				foundConfig.SetAcctNameAndType(acctName, credType)
				if hideSecret {
					foundConfig.SetSecret("*********")
				} else {
					passcode, passErr := rc.GetPassword(foundConfig.BuildCredPath(credType, acctName))
					if passErr != nil {
						ocelog.IncludeErrField(passErr).Error()
						foundConfig.SetSecret("ERROR: COULD NOT RETRIEVE PASSWORD FROM VAULT")
						err = passErr
					} else {
						foundConfig.SetSecret(passcode)
					}
				}
				creds[mapKey] = foundConfig
			}
			foundConfig.SetAdditionalFields(infoType, string(v.Value[:]))
		}
	} else {
		return creds, errors.New("not connected to consul, unable to retrieve credentials")
	}
	return creds, err
}

//CheckExists will tell you if a value exists at specified path
func (rc *RemoteConfig) CheckExists(path string) error {
	if rc.Consul.Connected {
		configs, err := rc.Consul.GetKeyValues(path)
		if err != nil {
			return err
		}
		for _, v := range configs {
			_, acctName, credType, infoType := splitConsulCredPath(v.Key)
			mapKey := credType + "/" + acctName
			foundConfig, ok := creds[mapKey]
			if !ok {
				foundConfig = instantiateCredObject(ocyType)
				foundConfig.SetAcctNameAndType(acctName, credType)
				if hideSecret {
					foundConfig.SetSecret("*********")
				} else {
					passcode, passErr := rc.GetPassword(foundConfig.BuildCredPath(credType, acctName))
					if passErr != nil {
						ocelog.IncludeErrField(passErr).Error()
						foundConfig.SetSecret("ERROR: COULD NOT RETRIEVE PASSWORD FROM VAULT")
						err = passErr
					} else {
						foundConfig.SetSecret(passcode)
					}
				}
				creds[mapKey] = foundConfig
			}
			foundConfig.SetAdditionalFields(infoType, string(v.Value[:]))
		}
	} else {
		return creds, errors.New("not connected to consul, unable to retrieve credentials")
	}
}

//GetPassword will return to you the vault password at specified path
func (rc *RemoteConfig) GetPassword(path string) (string, error) {
	authData, err := rc.Vault.GetUserAuthData(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", authData["clientsecret"]), nil
}

// AddRepoCreds adds repo integration creds to consul + vault
func (rc *RemoteConfig) AddCreds(path string, anyCred RemoteConfigCred) (err error) {
	if rc.Consul.Connected {
		anyCred.AddAdditionalFields(rc.Consul, path)
		if rc.Vault != nil {
			secret := make(map[string]interface{})
			secret["clientsecret"] = anyCred.GetClientSecret()
			if _, err = rc.Vault.AddUserAuthData(path, secret); err != nil {
				return
			}
		}
	} else {
		err = errors.New("not connected to consul, unable to add credentials")
	}
	return
}

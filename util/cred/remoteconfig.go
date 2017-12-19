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

// returns <vcs or repo>/acctname/credType/infoType
func splitConsulCredPath(path string) (typ OcyCredType, acctName, credType, infoType string) {
	pathKeys := strings.Split(path, "/")
	typ = OcyCredMap[pathKeys[1]]
	acctName = pathKeys[2]
	credType = pathKeys[3]
	infoType = pathKeys[4]
	return
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

// instantiateCredObject is what we will have to add too when we add new credential integrations
// (ie slack, w/e)
func instantiateCredObject(ocyType OcyCredType) models.Credential {
	switch ocyType {
	case Vcs:
		return &models.Credentials{}
	case Repo:
		return &models.RepoCreds{}
	default:
		panic("ahh!")
	}
}

// GetRepoCredAt will return repo credentials for repo integrations. its waaay to similar to getCredAt, so this is
// unacceptable. i can't figure out how to successfully use map[string]interface{}
func (remoteConfig *RemoteConfig) GetCredAt(path string, hideSecret bool, ocyType OcyCredType) (map[string]models.Credential, error) {
	creds := map[string]models.Credential{}
	var err error
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
				foundConfig = instantiateCredObject(ocyType)
				foundConfig.SetAcctNameAndType(acctName, credType)
				if hideSecret {
					foundConfig.SetSecret("*********")
				} else {
					passcode, passErr := remoteConfig.GetPassword(foundConfig.BuildCredPath(credType, acctName))
					//passcode, passErr := remoteConfig.GetPassword(ConfigPath + "/" + credType + "/" + acctName)
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

//GetPassword will return to you the vault password at specified path
func (remoteConfig *RemoteConfig) GetPassword(path string) (string, error) {
	authData, err := remoteConfig.Vault.GetUserAuthData(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", authData["clientsecret"]), nil
}

// AddRepoCreds adds repo integration creds to consul + vault
func (remoteConfig *RemoteConfig) AddCreds(path string, anyCred models.Credential) (err error) {
	if remoteConfig.Consul.Connected {
		anyCred.AddAdditionalFields(remoteConfig.Consul, path)
		if remoteConfig.Vault != nil {
			secret := make(map[string]interface{})
			secret["clientsecret"] = anyCred.GetClientSecret()
			if _, err = remoteConfig.Vault.AddUserAuthData(path, secret); err != nil {
				return
			}
		}
	} else {
		err = errors.New("not connected to consul, unable to add credentials")
	}
	return
}

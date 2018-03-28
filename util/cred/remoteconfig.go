package cred

import (
	"bitbucket.org/level11consulting/go-til/consul"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	ocevault "bitbucket.org/level11consulting/go-til/vault"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"fmt"
	"github.com/pkg/errors"
	"strconv"
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

type StorageCreds struct {
	User     string
	Location string
	Port     int
	DbName   string
	Password string
}

type StorageCred interface {
	GetStorageCreds(typ storage.Dest) (*StorageCreds, error)
	GetStorageType() (storage.Dest, error)
	GetOcelotStorage() (storage.OcelotStorage, error)
}


//CVRemoteConfig is an abstraction for retrieving/setting creds for ocelot
//currently uses consul + vault
type CVRemoteConfig interface {
	GetConsul()	*consul.Consulet
	SetConsul(consul *consul.Consulet)
	GetVault() ocevault.Vaulty
	SetVault(vault ocevault.Vaulty)
	GetCredAt(path string, hideSecret bool, rcc RemoteConfigCred) (map[string]RemoteConfigCred, error)
	GetPassword(path string) (string, error)
	AddCreds(path string, anyCred RemoteConfigCred) (err error)
	AddSSHKey(path string, sshKeyFile []byte) (err error)
	CheckSSHKeyExists(path string) (error)
	Healthy() bool
	Reconnect() error

	StorageCred
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

func (rc *RemoteConfig) Healthy() bool {
	vaultConnected := true
	_, err := rc.Vault.GetVaultData("here")
	if err != nil {
		if _, ok := err.(*ocevault.ErrNotFound); !ok {
			vaultConnected = false
		}
	}
	if !vaultConnected || !rc.Consul.Connected {
		ocelog.Log().Error("remoteConfig is not healthy")
		return false
	}
	return true
}

func (rc *RemoteConfig) Reconnect() error {
	_, err := rc.Vault.GetVaultData("here")
	if err != nil {
		if _, ok := err.(*ocevault.ErrNotFound); !ok {
			return err
		}
	}
	_, err = rc.Consul.GetKeyValue("here")
	if !rc.Consul.Connected {
		return err
	}
	return nil
}

// BuildCredKey returns the key for the map[string]RemoteConfigCred map that GetCredAt returns.
func BuildCredKey(credType string, acctName string) string {
	return credType + "/" + acctName
}

// GetCred at will return a map w/ key <cred_type>/<acct_name> to credentials. depending on the OcyCredType,
//   the appropriate credential struct will be instantiated and filled with data from consul and vault.
//   currently supports map[string]*models.VCSCreds and map[string]*models.RepoCreds
//   You must cast the resulting values to their appropriate objects after the map is generated if you need to access more than
//   the methods on the cred.RemoteConfigCred interface
//   Example:
//      creds, err := g.RemoteConfig.GetCredAt(cred.VCSPath, true, cred.Vcs)
//      vcsCreds := creds.(*models.VCSCreds)
func (rc *RemoteConfig) GetCredAt(path string, hideSecret bool, rcc RemoteConfigCred) (map[string]RemoteConfigCred, error) {
	creds := map[string]RemoteConfigCred{}
	var err error
	if rc.Consul.Connected {
		configs, err := rc.Consul.GetKeyValues(path)
		if err != nil {
			return creds, err
		}
		for _, v := range configs {
			_, acctName, credType, infoType := splitConsulCredPath(v.Key)
			mapKey := BuildCredKey(credType, acctName)
			foundConfig, ok := creds[mapKey]
			if !ok {
				foundConfig = rcc.Spawn()
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

// AddSSHKey adds repo ssh private key to vault at the usual vault path + /ssh
func (rc *RemoteConfig) AddSSHKey(path string, sshKeyFile []byte) (err error) {
	if rc.Vault != nil {
		secret := make(map[string]interface{})
		secret["sshKey"] = string(sshKeyFile)
		if _, err = rc.Vault.AddUserAuthData(path + "/ssh", secret); err != nil {
			return
		}
	} else {
		err = errors.New("no connection to vault, unable to add SSH Key")
	}
	return
}

// CheckSSHKey returns a boolean indicating whether or not an ssh key has been uploaded
func (rc *RemoteConfig) CheckSSHKeyExists(path string) (error) {
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


func (rc *RemoteConfig) GetStorageType() (storage.Dest, error) {
	kv, err := rc.Consul.GetKeyValue(StorageType)
	if err != nil {
		return 0, errors.New("unable to get storage type from consul, err: " + err.Error())
	}
	if kv == nil {
		ocelog.Log().Warning(fmt.Sprintf("there is no entry for storage type at the path \"%s\" in consul; using file system as the default.", StorageType))
		return storage.FileSystem, nil
	}
	storageType := string(kv.Value)
	switch storageType {
	case "postgres":
		return storage.Postgres, nil
	case "filesystem":
		return storage.FileSystem, nil
	default:
		return 0, errors.New("unknown storage type: " + storageType)
	}
}

func (rc *RemoteConfig) GetStorageCreds(typ storage.Dest) (*StorageCreds, error) {
	switch typ {
	case storage.Postgres:
		return rc.getForPostgres()
	case storage.FileSystem:
		return rc.getForFilesystem()
	default:
		fmt.Println("shouldnoteverhappen")
		return nil, nil
	}
}

func (rc *RemoteConfig) getForPostgres() (*StorageCreds, error) {
	pairs, err := rc.Consul.GetKeyValues(PostgresCredLoc)
	if err != nil {
		return nil, errors.New("unable to get postgres creds from consul, err: " + err.Error())
	}
	storeConfig := &StorageCreds{}
	for _, pair := range pairs {
		switch pair.Key {
		case PostgresDatabaseName:
			storeConfig.DbName = string(pair.Value)
		case PostgresLocation:
			storeConfig.Location = string(pair.Value)
		case PostgresUsername:
			storeConfig.User = string(pair.Value)
		case PostgresPort:
			// todo: check for err
			storeConfig.Port, _ = strconv.Atoi(string(pair.Value))
		}
	}
	secrets, err := rc.Vault.GetVaultData(PostgresPasswordLoc)
	if err != nil {
		return storeConfig, errors.New("unable to get postgres password from vault, err: " + err.Error())
	}
	// making name clientsecret because i feel like there must be a way for us to genericize remoteConfig
	storeConfig.Password = fmt.Sprintf("%v", secrets[PostgresPasswordKey])
	return storeConfig, nil
}

func (rc *RemoteConfig) getForFilesystem() (*StorageCreds, error) {
	pair, err := rc.Consul.GetKeyValue(FilesystemDir)
	if err != nil {
		return nil, errors.New("unable to get save directory from consul, err: " + err.Error())
	}
	return &StorageCreds{Location: string(pair.Value)}, nil
}

func (rc *RemoteConfig) GetOcelotStorage() (storage.OcelotStorage, error) {
	typ, err := rc.GetStorageType()
	if err != nil {
		return nil, err
	}
	if typ == storage.Postgres {
		fmt.Println("postgres storage")
	}
	creds, err := rc.GetStorageCreds(typ)
	if err != nil {
		return nil, err
	}
	switch typ {
	case storage.FileSystem:
		return storage.NewFileBuildStorage(creds.Location), nil
	case storage.Postgres:
		return storage.NewPostgresStorage(creds.User, creds.Password, creds.Location, creds.Port, creds.DbName), nil
	default:
		return nil, errors.New("unknown type")
	}
	return nil, errors.New("could not grab ocelot storage")
}


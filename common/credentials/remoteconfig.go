package credentials

import (
	"fmt"
	"strconv"
	"bitbucket.org/level11consulting/go-til/consul"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	ocevault "bitbucket.org/level11consulting/go-til/vault"
	pb "bitbucket.org/level11consulting/ocelot/old/admin/models"
	"bitbucket.org/level11consulting/ocelot/newocy/integrations"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"github.com/pkg/errors"
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

type HealthyMaintainer interface {
	Reconnect() error
	Healthy() bool
}

//CVRemoteConfig is an abstraction for retrieving/setting creds for ocelot
//currently uses consul + vault
type CVRemoteConfig interface {
	GetConsul()	*consul.Consulet
	SetConsul(consul *consul.Consulet)
	GetVault() ocevault.Vaulty
	SetVault(vault ocevault.Vaulty)
	AddSSHKey(path string, sshKeyFile []byte) (err error)
	CheckSSHKeyExists(path string) (error)
	GetPassword(scType pb.SubCredType, acctName string, ocyCredType pb.CredType, identifier string) (string, error)
	InsecureCredStorage
	HealthyMaintainer

	StorageCred
}

type InsecureCredStorage interface {
	GetCredsByType(store storage.CredTable, ctype pb.CredType, hideSecret bool) ([]pb.OcyCredder, error)
	GetAllCreds(store storage.CredTable, hideSecret bool) ([]pb.OcyCredder, error)
	GetCred(store storage.CredTable, subCredType pb.SubCredType, identifier, accountName string, hideSecret bool) (pb.OcyCredder, error)
	GetCredsBySubTypeAndAcct(store storage.CredTable, stype pb.SubCredType, accountName string, hideSecret bool) ([]pb.OcyCredder, error)
	AddCreds(store storage.CredTable, anyCred pb.OcyCredder, overwriteOk bool) (err error)
	UpdateCreds(store storage.CredTable, anyCred pb.OcyCredder) (err error)
}

type NewCVRC interface {

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

// todo: write a test for thiiiiis!
func (rc *RemoteConfig) Healthy() bool {
	vaultConnected := true
	_, err := rc.Vault.GetVaultData("here")
	if err != nil {
		if _, ok := err.(*ocevault.ErrNotFound); !ok {
			vaultConnected = false
		}
	}
	rc.Consul.GetKeyValue("here")
	if !vaultConnected || !rc.Consul.Connected  {
		ocelog.Log().Error("remoteConfig is not healthy")
		return false
	}
	return true
}

//todo: write a test for this!!!
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


// AddSSHKey adds repo ssh private key to vault at the usual vault path + /ssh
func (rc *RemoteConfig) AddSSHKey(path string, sshKeyFile []byte) (err error) {
	if rc.Vault != nil {
		secret := buildSecretPayload(string(sshKeyFile))
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
func (rc *RemoteConfig) GetPassword(scType pb.SubCredType, acctName string, ocyCredType pb.CredType, identifier string) (string, error) {
	authData, err := rc.Vault.GetUserAuthData(BuildCredPath(scType, acctName, ocyCredType, identifier))
	if err != nil {
		return "", err
	}
	secretData := authData["data"]
	password, ok := secretData.(map[string]interface{})
	if !ok {
		return "", errors.New("Could not retrieve password from vault") //how is it that we can't cast to a map of string/string??
	}
	passwordStr := password["clientsecret"].(string)
	return passwordStr, nil
}

// AddRepoCreds adds repo integration creds to storage + vault
func (rc *RemoteConfig) AddCreds(store storage.CredTable, anyCred pb.OcyCredder, overwriteOk bool) (err error) {
	if rc.Vault != nil {
		path := BuildCredPath(anyCred.GetSubType(), anyCred.GetAcctName(), anyCred.GetSubType().Parent(), anyCred.GetIdentifier())

		dataWrapper := buildSecretPayload(anyCred.GetClientSecret())
		if _, err = rc.Vault.AddUserAuthData(path, dataWrapper); err != nil {
			return
		}
	} else {
		return errors.New("remote config not properly initialized, cannot add creds")
	}
	if err := store.InsertCred(anyCred, overwriteOk); err != nil {
		return err
	}
	return
}

func (rc *RemoteConfig) UpdateCreds(store storage.CredTable, anyCred pb.OcyCredder) (err error) {
	if rc.Vault != nil {
		path := BuildCredPath(anyCred.GetSubType(), anyCred.GetAcctName(), anyCred.GetSubType().Parent(), anyCred.GetIdentifier())

		dataWrapper := buildSecretPayload(anyCred.GetClientSecret())
		if _, err = rc.Vault.AddUserAuthData(path, dataWrapper); err != nil {
			return
		}
	} else {
		return errors.New("remote config not properly initialized, cannot add creds")
	}
	err = store.UpdateCred(anyCred)
	return
}


//this builds the secret payload as accepted by vault docs here: https://www.vaultproject.io/api/secret/kv/kv-v2.html
func buildSecretPayload(secret string) map[string]interface{} {
	dataWrapper := make(map[string]interface{})
	topSecret := make(map[string]string)
	topSecret["clientsecret"] = secret
	dataWrapper["data"] = topSecret
	return dataWrapper
}



func (rc *RemoteConfig) maybeGetPassword(subCredType pb.SubCredType, accountName string, hideSecret bool, identifier string) (secret string){
	if !hideSecret {
		passcode, passErr := rc.GetPassword(subCredType, accountName, subCredType.Parent(), identifier)
		if passErr != nil {
			ocelog.IncludeErrField(passErr).Error()
			secret = "ERROR: COULD NOT RETRIEVE PASSWORD FROM VAULT"
		} else {
			secret = passcode
		}
	} else {
		secret = "*********"
	}
	return secret
}

func (rc *RemoteConfig) GetCred(store storage.CredTable, subCredType pb.SubCredType, identifier, accountName string, hideSecret bool) (pb.OcyCredder, error) {
	cred, err := store.RetrieveCred(subCredType, identifier, accountName)
	if err != nil {
		return nil, err
	}
	cred.SetSecret(rc.maybeGetPassword(subCredType, accountName, hideSecret, identifier))
	return cred, err
}

func (rc *RemoteConfig) GetAllCreds(store storage.CredTable, hideSecret bool) ([]pb.OcyCredder, error) {
	creds, err := store.RetrieveAllCreds()
	if err != nil {
		return creds, err
	}
	var allcreds []pb.OcyCredder
	for _, cred := range creds {
		sec := rc.maybeGetPassword(cred.GetSubType(), cred.GetAcctName(), hideSecret, cred.GetIdentifier())
		cred.SetSecret(sec)
		allcreds = append(allcreds, cred)
	}
	return allcreds, nil
}

func (rc *RemoteConfig) GetCredsByType(store storage.CredTable, ctype pb.CredType, hideSecret bool) ([]pb.OcyCredder, error) {
	creds, err := store.RetrieveCreds(ctype)
	if err != nil {
		return creds, err
	}
	var credsfortype []pb.OcyCredder
	for _, cred := range creds {
		sec := rc.maybeGetPassword(cred.GetSubType(), cred.GetAcctName(), hideSecret, cred.GetIdentifier())
		cred.SetSecret(sec)
		credsfortype = append(credsfortype, cred)
	}
	return credsfortype, nil
}

func (rc *RemoteConfig) GetCredsBySubTypeAndAcct(store storage.CredTable, stype pb.SubCredType, accountName string, hideSecret bool) ([]pb.OcyCredder, error) {
	creds, err := store.RetrieveCredBySubTypeAndAcct(stype, accountName)
	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); ok {
			return nil, integrations.NCErr(fmt.Sprintf("credentials not found for account %s and integration %s", accountName, "kubeconf"))
		}
		return nil, err
	}
	var credsForType []pb.OcyCredder
	for _, cred := range creds {
		sec := rc.maybeGetPassword(stype, accountName, hideSecret, cred.GetIdentifier())
		cred.SetSecret(sec)
		credsForType = append(credsForType, cred)
	}
	return credsForType, nil
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


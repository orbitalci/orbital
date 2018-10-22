package credentials

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shankj3/go-til/consul"
	ocelog "github.com/shankj3/go-til/log"
	ocevault "github.com/shankj3/go-til/vault"
	"github.com/shankj3/ocelot/common"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"
)

var (
	failedCredRetrieval = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ocelot_failed_cred",
			Help: "number of times ocelot is unable to interact with cred store",
		},
		// interaction_type: create | read | update | delete
		// cred_type: subcredType.String()
		// is_secret: bool , whether or not interacting with secret store or insecure cred store
		[]string{"cred_type", "interaction_type", "is_secret"},
	)
)

func init() {
	prometheus.MustRegister(failedCredRetrieval)
}

//go:generate mockgen -source remoteconfig.go -destination remoteconfig.mock.go -package credentials

// GetToken will check for a vault token first in the environment variable VAULT_TOKEN. If not found at the env var,
// either the path given or the default path of /etc/vaulted/token will be searched to see if it exists. If it exists,
// its contents will be read and returned as the vault token. for kubernetes support
func GetToken(path string) (string, error) {
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

//GetInstance returns a new instance of ConfigConsult. If consulHot and consulPort are empty,
//this will talk to consul using reasonable defaults (localhost:8500)
// FIXME: This condense consulHost and consulPort into a single connection string
//if token is an empty string, vault will be initialized with $VAULT_TOKEN
func GetInstance(consulHost string, consulPort int, token string) (CVRemoteConfig, error) {
	remoteConfig := &RemoteConfig{}

	// FIXME: If we are reading from CONSUL_HTTP_ADDR, we may encounter Unix socket paths. For now, let's pretend our perfect world is always http

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
		// right now, only search for token at default path
		var err error
		token, err = GetToken("")

		if err != nil {
			return nil, err
		}
	}
	vaultClient, err := ocevault.NewAuthedClient(token)
	if err != nil {
		return nil, err
	}
	remoteConfig.Vault = vaultClient

	return remoteConfig, nil
}

type StorageCreds struct {
	User         string
	Location     string
	Port         int
	DbName       string
	Password     string
	VaultLeaseID string
}

type StorageCred interface {
	GetStorageCreds(typ *storage.Dest) (*StorageCreds, error)
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
	GetConsul() consul.Consuletty
	SetConsul(consul consul.Consuletty)
	GetVault() ocevault.Vaulty
	SetVault(vault ocevault.Vaulty)
	AddSSHKey(path string, sshKeyFile []byte) (err error)
	CheckSSHKeyExists(path string) error
	GetPassword(scType pb.SubCredType, acctName string, ocyCredType pb.CredType, identifier string) (string, error)
	DeleteCred(store storage.CredTable, anyCred pb.OcyCredder) (err error)
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

//
//type CredStore struct {
//	RC    CVRemoteConfig
//	Store storage.CredTable
//}

// RemoteConfig returns a struct with client handlers for Consul and Vault. Mainly for passing around after authenticating
type RemoteConfig struct {
	Consul consul.Consuletty
	Vault  ocevault.Vaulty
}

// GetConsul returns the local consul client handler
func (rc *RemoteConfig) GetConsul() consul.Consuletty {
	return rc.Consul
}

// SetConsul sets the local consul client handler
func (rc *RemoteConfig) SetConsul(consl consul.Consuletty) {
	rc.Consul = consl
}

// GetVault returns the local vault client handler
func (rc *RemoteConfig) GetVault() ocevault.Vaulty {
	return rc.Vault
}

// SetVault sets the local vault client handler
func (rc *RemoteConfig) SetVault(vault ocevault.Vaulty) {
	rc.Vault = vault
}

func (rc *RemoteConfig) Healthy() bool {
	vaultConnected := true
	_, err := rc.Vault.GetVaultData("secret/data/config/ocelot/here")
	if err != nil {
		if _, ok := err.(*ocevault.ErrNotFound); !ok {
			vaultConnected = false
		}
	}
	rc.Consul.GetKeyValue("here")
	if !vaultConnected || !rc.Consul.IsConnected() {
		ocelog.Log().Error("remoteConfig is not healthy")
		return false
	}
	return true
}

func (rc *RemoteConfig) Reconnect() error {
	_, err := rc.Vault.GetVaultData("secret/data/config/ocelot/here")
	if err != nil {
		if _, ok := err.(*ocevault.ErrNotFound); !ok {
			return err
		}
	}
	_, err = rc.Consul.GetKeyValue("here")
	if !rc.Consul.IsConnected() {
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

func (rc *RemoteConfig) deletePassword(scType pb.SubCredType, acctName, identifier string) error {
	credPath := BuildCredPath(scType, acctName, scType.Parent(), identifier)
	ocelog.Log().Debug("CREDPATH=", credPath)
	if err := rc.Vault.DeletePath(credPath); err != nil {
		return errors.WithMessage(err, "Unable to delete password for user "+acctName+" w/ identifier "+identifier)
	}
	return nil
}

func (rc *RemoteConfig) DeleteCred(store storage.CredTable, anyCred pb.OcyCredder) (err error) {
	if storeErr := store.DeleteCred(anyCred); storeErr != nil {
		err = errors.WithMessage(storeErr, "unable to delete un-sensitive data")
	}
	if secureErr := rc.deletePassword(anyCred.GetSubType(), anyCred.GetAcctName(), anyCred.GetIdentifier()); secureErr != nil {

		err2 := errors.WithMessage(secureErr, "unable to delete sensitive data ")
		if err == nil {
			err = err2
		} else {
			err = errors.Wrap(err, err2.Error())
		}
	}
	return err
}

//GetPassword will return to you the vault password at specified path
func (rc *RemoteConfig) GetPassword(scType pb.SubCredType, acctName string, ocyCredType pb.CredType, identifier string) (string, error) {
	credPath := BuildCredPath(scType, acctName, ocyCredType, identifier)
	ocelog.Log().Debug("CREDPATH=", credPath)
	authData, err := rc.Vault.GetUserAuthData(credPath)
	if err != nil {
		failedCredRetrieval.WithLabelValues(scType.String(), "read", "true")
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

// AddCreds adds repo integration creds to storage + vault
func (rc *RemoteConfig) AddCreds(store storage.CredTable, anyCred pb.OcyCredder, overwriteOk bool) (err error) {
	if rc.Vault != nil {
		path := BuildCredPath(anyCred.GetSubType(), anyCred.GetAcctName(), anyCred.GetSubType().Parent(), anyCred.GetIdentifier())

		dataWrapper := buildSecretPayload(anyCred.GetClientSecret())
		if _, err = rc.Vault.AddUserAuthData(path, dataWrapper); err != nil {
			failedCredRetrieval.WithLabelValues(anyCred.GetSubType().String(), "create", "true")
			return
		}
	} else {
		return errors.New("remote config not properly initialized, cannot add creds")
	}
	if err := store.InsertCred(anyCred, overwriteOk); err != nil {
		failedCredRetrieval.WithLabelValues(anyCred.GetSubType().String(), "create", "false")
		return err
	}
	return
}

func (rc *RemoteConfig) UpdateCreds(store storage.CredTable, anyCred pb.OcyCredder) (err error) {
	if rc.Vault != nil {
		path := BuildCredPath(anyCred.GetSubType(), anyCred.GetAcctName(), anyCred.GetSubType().Parent(), anyCred.GetIdentifier())

		dataWrapper := buildSecretPayload(anyCred.GetClientSecret())
		if _, err = rc.Vault.AddUserAuthData(path, dataWrapper); err != nil {
			failedCredRetrieval.WithLabelValues(anyCred.GetSubType().String(), "update", "true")
			return
		}
	} else {
		return errors.New("remote config not properly initialized, cannot add creds")
	}
	err = store.UpdateCred(anyCred)
	if err != nil {
		failedCredRetrieval.WithLabelValues(anyCred.GetSubType().String(), "update", "false")
	}
	return
}

//buildSecretPayload builds the secret payload as accepted by vault docs here: https://www.vaultproject.io/api/secret/kv/kv-v2.html
func buildSecretPayload(secret string) map[string]interface{} {
	dataWrapper := make(map[string]interface{})
	topSecret := make(map[string]string)
	topSecret["clientsecret"] = secret
	dataWrapper["data"] = topSecret
	return dataWrapper
}

func (rc *RemoteConfig) maybeGetPassword(subCredType pb.SubCredType, accountName string, hideSecret bool, identifier string) (secret string) {
	if !hideSecret {
		passcode, passErr := rc.GetPassword(subCredType, accountName, subCredType.Parent(), identifier)
		if passErr != nil {
			failedCredRetrieval.WithLabelValues(subCredType.String(), "read", "true")
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
		failedCredRetrieval.WithLabelValues(subCredType.String(), "read", "false")
		return nil, err
	}
	cred.SetSecret(rc.maybeGetPassword(subCredType, accountName, hideSecret, identifier))
	return cred, err
}

func (rc *RemoteConfig) GetAllCreds(store storage.CredTable, hideSecret bool) ([]pb.OcyCredder, error) {
	creds, err := store.RetrieveAllCreds()
	if err != nil {
		failedCredRetrieval.WithLabelValues("ALL", "read", "false")
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
			return nil, common.NCErr(fmt.Sprintf("credentials not found for account %s and integration %s", accountName, "kubeconf"))
		}
		failedCredRetrieval.WithLabelValues(stype.String(), "read", "false")
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
	kv, err := rc.Consul.GetKeyValue(common.StorageType)
	if err != nil {
		return 0, errors.New("unable to get storage type from consul, err: " + err.Error())
	}
	if kv == nil {
		ocelog.Log().Warning(fmt.Sprintf("there is no entry for storage type at the path \"%s\" in consul; using file system as the default.", common.StorageType))
		return storage.FileSystem, nil
	}

	// ?: Is there an overall positive experience for making this value case-insensitive?
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

// GetStorageCreds initializes datastore info based on the configured storage type in Consul
// FIXME: We're doing ourselves a disservice by forcing a user to pass in a *storage.Dest when we can handle this internally through rc
func (rc *RemoteConfig) GetStorageCreds(typ *storage.Dest) (*StorageCreds, error) {
	switch *typ {
	case storage.Postgres:
		return rc.getForPostgres()
	case storage.FileSystem:
		return rc.getForFilesystem()
	default:
		return nil, errors.New("Failed to get storage creds for Postgres or Filesystem. This shouldn't ever happen, but it did")
	}
}

// getForPostgres Reads configuration from Consul
func (rc *RemoteConfig) getForPostgres() (*StorageCreds, error) {

	// Credentials for Postgres may come from 2 mutually exclusive areas in Vault
	// Either creds are static, and stored as a KV secret in vault
	// Or creds are dynamic, and managed by vault's database secret engine

	// The desired default behavior is to use static secrets. Don't error out if Vault settings aren't in Consul
	vaultConf, _ := rc.Consul.GetKeyValues(common.VaultConf)

	vaultBackend := "kv" // The default backend is "kv" if left unconfigured in Consul
	for _, vconf := range vaultConf {
		switch vconf.Key {
		case common.VaultDBSecretEngine:
			switch string(vconf.Value) {
			case "kv":
				ocelog.Log().Info("Static Postgres creds")
				break
			case "database":
				ocelog.Log().Info("Dynamic Postgres creds")
				break
			default:
				return &StorageCreds{}, errors.New("Unsupported Vault DB Secret Engine")
			}
			vaultBackend = string(vconf.Value)
		}
	}

	storeConfig := &StorageCreds{}

	kvconfig, err := rc.Consul.GetKeyValues(common.PostgresCredLoc)
	if len(kvconfig) == 0 || err != nil {
		errorMsg := fmt.Sprintf("unable to get postgres creds from consul")
		if err != nil {
			errorMsg = fmt.Sprintf("%s, err: %s", errorMsg, err.Error())
		}
		return nil, errors.New(errorMsg)
	}

	for _, pair := range kvconfig {
		switch pair.Key {
		case common.PostgresDatabaseName:
			storeConfig.DbName = string(pair.Value)
		case common.PostgresLocation:
			storeConfig.Location = string(pair.Value)
		case common.PostgresPort:
			// todo: check for err
			storeConfig.Port, _ = strconv.Atoi(string(pair.Value))
		}
	}

	switch vaultBackend {
	case "kv":
		kvconfig, err := rc.Consul.GetKeyValues(common.PostgresCredLoc)
		if len(kvconfig) == 0 || err != nil {
			errorMsg := fmt.Sprintf("unable to get postgres location from consul")
			if err != nil {
				errorMsg = fmt.Sprintf("%s, err: %s", errorMsg, err.Error())
			}
			return &StorageCreds{}, errors.New(errorMsg)
		}
		for _, pair := range kvconfig {
			switch pair.Key {
			case common.PostgresUsername:
				storeConfig.User = string(pair.Value)
			}
		}

		secrets, err := rc.Vault.GetVaultData(common.PostgresPasswordLoc)
		if len(secrets) == 0 || err != nil {
			errorMsg := fmt.Sprintf("unable to get postgres password from consul")
			if err != nil {
				errorMsg = fmt.Sprintf("%s, err: %s", errorMsg, err.Error())
			}
			return &StorageCreds{}, errors.New(errorMsg)
		}

		// making name clientsecret because i feel like there must be a way for us to genericize remoteConfig
		storeConfig.Password = fmt.Sprintf("%v", secrets[common.PostgresPasswordKey])

	case "database":
		secrets, err := rc.Vault.GetVaultSecret("database/creds/ocelot")
		if err != nil {
			errorMsg := fmt.Sprintf("unable to get dynamic postgres creds from vault")
			if err != nil {
				errorMsg = fmt.Sprintf("%s, err: %s", errorMsg, err.Error())
			}
			return &StorageCreds{}, errors.New(errorMsg)
		}

		storeConfig.User = fmt.Sprintf("%v", secrets.Data["username"].(string))
		storeConfig.Password = fmt.Sprintf("%v", secrets.Data["password"].(string))

		ocelog.Log().Debugf("Dynamic postgres creds from Vault: %v", secrets)

		// Kick of a lease renewal goroutine
		go rc.Vault.RenewLeaseForever(secrets)
	}

	return storeConfig, nil
}

func (rc *RemoteConfig) getForFilesystem() (*StorageCreds, error) {
	pair, err := rc.Consul.GetKeyValue(common.FilesystemDir)

	if err != nil {
		return nil, errors.New("unable to get save directory from consul, err: " + err.Error())
	}
	return &StorageCreds{Location: string(pair.Value)}, nil
}

// GetOcelotStorage instantiates the datastore based on Consul configuration for Ocelot. Opens the database connection for Postgres.
func (rc *RemoteConfig) GetOcelotStorage() (storage.OcelotStorage, error) {
	// FIXME: We should get rid of this call, and let GetStorageCreds() handle more of this
	typ, err := rc.GetStorageType()
	if err != nil {
		return nil, err
	}
	if typ == storage.Postgres {
		fmt.Println("postgres storage")
	}

	// This might return the full secrets struct
	creds, err := rc.GetStorageCreds(&typ)
	if err != nil {
		return nil, err
	}

	/// Can I just pass creds? This would be more convenient
	switch typ {
	case storage.FileSystem:
		return storage.NewFileBuildStorage(creds.Location), nil
	case storage.Postgres:
		store := storage.NewPostgresStorage(creds.User, creds.Password, creds.Location, creds.Port, creds.DbName)
		//ocelog.Log().Debugf("user %s pw %s loc %s port %s db %s", creds.User, creds.Password, creds.Location, creds.Port, creds.DbName)

		return store, store.Connect()
	default:
		return nil, errors.New("unknown type")
	}
	return nil, errors.New("could not grab ocelot storage")
}

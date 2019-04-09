package config

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/level11consulting/ocelot/common"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/storage"
	"github.com/level11consulting/ocelot/storage/postgres"
	"github.com/pkg/errors"
	"github.com/shankj3/go-til/consul"
	ocelog "github.com/shankj3/go-til/log"
	ocevault "github.com/shankj3/go-til/vault"
	"github.com/level11consulting/ocelot/common/credentials"
)

//GetInstance attempts to connect to Consul and Vault, returns a new instance remoteConfig
func GetInstance(consulURI *url.URL, vaultToken string) (CVRemoteConfig, error) {
	remoteConfig := &RemoteConfig{}

	//intialize consul
	var portInt, _ = strconv.Atoi(consulURI.Port())
	if consulURI.Hostname() == "" && portInt == 0 {
		consulet, err := consul.Default()
		if err != nil {
			return nil, err
		}
		remoteConfig.Consul = consulet
	} else {

		// This is certainly an area to refactor in the future
		consulet, err := consul.New(consulURI.Hostname(), portInt)
		remoteConfig.Consul = consulet
		if err != nil {
			return nil, err
		}
	}

	//initialize vault
	vaultToken, err := getVaultToken(vaultToken)

	if err != nil {
		return nil, err
	}

	vaultClient, err := ocevault.NewAuthedClient(vaultToken)
	if err != nil {
		return nil, err
	}
	remoteConfig.Vault = vaultClient

	return remoteConfig, nil
}

// RemoteConfig returns a struct with client handlers for Consul and Vault. Mainly for passing around after authenticating
type RemoteConfig struct {
	Consul consul.Consuletty
	Vault  ocevault.Vaulty
}

// Healthy returns the status of whether Vault and Consul are currently connected
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

// buildCredKey returns the key for the map[string]RemoteConfigCred map that GetCredAt returns.
func buildCredKey(credType string, acctName string) string {
	return credType + "/" + acctName
}

func (rc *RemoteConfig) deletePassword(scType pb.SubCredType, acctName, identifier string) error {
	credPath := credentials.BuildCredPath(scType, acctName, scType.Parent(), identifier)
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

// FIXME: For consistency, this should be renamed to something like GetCred, but unique... Make it clear what is a Cred and a Password
//GetPassword will return to you the vault password at specified path
func (rc *RemoteConfig) GetPassword(scType pb.SubCredType, acctName string, ocyCredType pb.CredType, identifier string) (string, error) {
	credPath := credentials.BuildCredPath(scType, acctName, ocyCredType, identifier)
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
		path := credentials.BuildCredPath(anyCred.GetSubType(), anyCred.GetAcctName(), anyCred.GetSubType().Parent(), anyCred.GetIdentifier())

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
		path := credentials.BuildCredPath(anyCred.GetSubType(), anyCred.GetAcctName(), anyCred.GetSubType().Parent(), anyCred.GetIdentifier())

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

// GetStorageType reads from consul at common.StorageType, and returns a handle for the configured storage.
func (rc *RemoteConfig) GetStorageType() (storage.Dest, error) {
	kv, err := rc.Consul.GetKeyValue(common.StorageType)
	if err != nil {
		return 0, errors.New("unable to get storage type from consul, err: " + err.Error())
	}
	if kv == nil {
		return 0, errors.Errorf("there is no entry for storage type at the path \"%s\" in consul, required", common.StorageType)
	}
	// ?: Is there an overall positive experience for making this value case-insensitive?
	storageType := string(kv.Value)
	switch storageType {
	case "postgres":
		return storage.Postgres, nil
	default:
		return 0, errors.New("unknown storage type: " + storageType)
	}
}

// FIXME: We're doing ourselves a disservice by forcing a user to pass in a *storage.Dest when we can handle this internally through rc
// GetStorageCreds initializes datastore info based on the configured storage type in Consul
func (rc *RemoteConfig) GetStorageCreds(typ *storage.Dest) (*StorageCreds, error) {
	switch *typ {
	case storage.Postgres:
		return rc.getForPostgres()
	default:
		return nil, errors.New("Failed to get storage creds for Postgres or Filesystem. This shouldn't ever happen, but it did")
	}
}

/////// We want to clean up the calls that use/switch on typ, and creds.
/////// The new storage should be cleaned up to take storage.OcelotStorage, instead of the elements within

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
	case storage.Postgres:
		store, _ := storage_postgres.NewPostgresStorage(creds.User, creds.Password, creds.Location, creds.Port, creds.DbName)
		//ocelog.Log().Debugf("user %s pw %s loc %s port %s db %s", creds.User, creds.Password, creds.Location, creds.Port, creds.DbName)

		return store, store.Connect()
	default:
		return nil, errors.New("unknown type")
	}
	return nil, errors.New("could not grab ocelot storage. This error should be unreachable")
}

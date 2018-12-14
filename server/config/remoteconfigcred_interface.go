package config

import (
	"github.com/level11consulting/orbitalci/models/pb"
	"github.com/level11consulting/orbitalci/storage"
	"github.com/shankj3/go-til/consul"
	ocevault "github.com/shankj3/go-til/vault"
)

//go:generate mockgen -source remoteconfigcred_interface.go -destination remoteconfigcred_interface.mock.go -package credentials

// RemoteConfigCred is the interface that remoteConfig requires credential structs to adhere to
// to appropriately add things to consul and vault. Implementation can be seen in `admin/models/guideocelotmodels.go`.
type RemoteConfigCred interface {
	GetClientSecret() string
	SetAcctNameAndType(name string, typ string)
	GetAcctName() string
	GetType() string
	SetSecret(string)
	SetAdditionalFields(key string, val string)
	AddAdditionalFields(consule *consul.Consulet, path string) error
	BuildCredPath(credType string, acctName string) string
	Spawn() RemoteConfigCred
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

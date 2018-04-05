package cred

import (
	"bitbucket.org/level11consulting/go-til/consul"
)

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
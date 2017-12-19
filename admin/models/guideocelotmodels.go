package models

import (
	"bitbucket.org/level11consulting/go-til/consul"
	"fmt"
)

// interface that proto objects have to fulfill to be able to log their credentials in vault / consul
// you can see how it is used in creds/remoteconfig.go
type Credential interface {
	GetClientSecret() string
	SetAcctNameAndType(name string, typ string)
	SetSecret(string)
	SetAdditionalFields(key string, val string)
	AddAdditionalFields(consule *consul.Consulet, path string) error
	BuildCredPath(credType string, acctName string) string
}

// these methods are attached to the proto object RepoCreds
func (repoCred *RepoCreds) SetAcctNameAndType(name string, typ string) {
	repoCred.AcctName = name
	repoCred.Type = typ
}

func (repoCred *RepoCreds) BuildCredPath(credType string, acctName string) string {
	return fmt.Sprintf("%s/repo/%s/%s", "creds", acctName, credType)
}

func (repoCred *RepoCreds) SetSecret(secret string) {
	repoCred.Password = secret
}

func (repoCred *RepoCreds) GetClientSecret() string {
	return repoCred.Password
}

func (repoCred *RepoCreds) SetAdditionalFields(infoType string, val string) {
	switch infoType {
	case "repourl":
		repoCred.RepoUrl = val
	case "username":
		repoCred.Username = val
	}
}

func (repoCred *RepoCreds) AddAdditionalFields(consule *consul.Consulet, path string) (err error) {
	if err := consule.AddKeyValue(path + "/username", []byte(repoCred.Username)); err != nil {
		return err
	}
	if err = consule.AddKeyValue(path + "/repourl", []byte(repoCred.RepoUrl)); err != nil {
		return err
	}
	return err
}


// these methods are to enable remoteconfig cred save with the proto Credentials object
func (adminCred *Credentials) SetAcctNameAndType(name string, typ string) {
	adminCred.AcctName = name
	adminCred.Type = typ
}

func (adminCred *Credentials) BuildCredPath(credType string, acctName string) string {
	return fmt.Sprintf("%s/vcs/%s/%s", "creds", acctName, credType)
}

func (adminCred *Credentials) SetSecret(secret string) {
	adminCred.ClientSecret = secret
}

func (adminCred *Credentials) SetAdditionalFields(infoType string, val string) {
	switch infoType {
	case "clientid":
		adminCred.ClientId = val
	case "tokenurl":
		adminCred.TokenURL = val
	}
}

func (adminCred *Credentials) AddAdditionalFields(consule *consul.Consulet, path string) error {
	err := consule.AddKeyValue(path+"/clientid", []byte(adminCred.ClientId))
	if err != nil {
		return err
	}
	err = consule.AddKeyValue(path+"/tokenurl", []byte(adminCred.TokenURL))
	if err != nil {
		return err
	}
	return err
}
package cred

import (
	"bitbucket.org/level11consulting/go-til/consul"
	"fmt"
	"strings"
)

//dummy vcscred obj so we don't have to depend on models anymore
type VcsConfig struct {
	ClientId     string
	ClientSecret string
	TokenURL     string
	AcctName     string
	Type         string
}

func (m *VcsConfig) GetClientSecret() string {
	return m.ClientSecret
}

func (m *VcsConfig) SetAcctNameAndType(name string, typ string) {
	m.AcctName = name
	m.Type = typ
}

func (m *VcsConfig) BuildCredPath(credType string, acctName string) string {
	return fmt.Sprintf("%s/vcs/%s/%s", "creds", acctName, credType)
}

func (m *VcsConfig) SetSecret(secret string) {
	m.ClientSecret = secret
}

func (m *VcsConfig) SetAdditionalFields(infoType string, val string) {
	switch infoType {
	case "clientid":
		m.ClientId = val
	case "tokenurl":
		m.TokenURL = val
	}
}

func (m *VcsConfig) AddAdditionalFields(consule *consul.Consulet, path string) error {
	err := consule.AddKeyValue(path+"/clientid", []byte(m.ClientId))
	if err != nil {
		return err
	}
	err = consule.AddKeyValue(path+"/tokenurl", []byte(m.TokenURL))
	if err != nil {
		return err
	}
	return err
}

func (m *VcsConfig) Spawn() RemoteConfigCred {
	return &VcsConfig{}
}


type RepoConfig struct {
	Username    string
	Password    string
	RepoUrl     map[string]string
	AcctName    string
	Type        string
	ProjectName string
}


// these methods are attached to the proto object RepoConfig
func (m *RepoConfig) SetAcctNameAndType(name string, typ string) {
	m.AcctName = name
	m.Type = typ
}

func (m *RepoConfig) BuildCredPath(credType string, acctName string) string {
	return fmt.Sprintf("%s/repo/%s/%s", "creds", acctName, credType)
}

func (m *RepoConfig) SetSecret(secret string) {
	m.Password = secret
}

func (m *RepoConfig) GetClientSecret() string {
	return m.Password
}

func (m *RepoConfig) SetAdditionalFields(infoType string, val string) {
	if strings.Contains(infoType, "repourl") {
		paths := strings.Split(infoType, "/")
		if len(paths) > 2 {
			panic("WHAT THE FUCK?")
		}
		m.RepoUrl[paths[1]] = val
	}
	if infoType == "username" {
		m.Username = val
	}
}

func (m *RepoConfig) AddAdditionalFields(consule *consul.Consulet, path string) (err error) {
	if err := consule.AddKeyValue(path + "/username", []byte(m.Username)); err != nil {
		return err
	}
	for reponame, url := range m.RepoUrl {
		if err = consule.AddKeyValue(path + "/repourl/" + reponame, []byte(url)); err != nil {
			return err
		}
	}
	return err
}

func (m *RepoConfig) Spawn() RemoteConfigCred {
	return &RepoConfig{RepoUrl: make(map[string]string)}
}

//if shnak.GetPassword() != repoCreds.GetPassword() {
//t.Error(test.StrFormatErrors("repo password", repoCreds.Password, shnak.Password))
//}
//if shnak.GetType() != repoCreds.GetType() {
//t.Error(test.StrFormatErrors("repo acct type", repoCreds.GetType(), shnak.GetType()))
//}
//if shnak.GetRepoUrl() != repoCreds.GetRepoUrl() {
//t.Error(test.StrFormatErrors("repo url", repoCreds.GetRepoUrl(), shnak.GetRepoUrl()))
//}

func (m *RepoConfig) GetPassword() string {
	return m.Password
}

func (m *RepoConfig) GetType() string {
	return m.Type
}


func (m *RepoConfig) GetRepoUrl() map[string]string {
	return m.RepoUrl
}
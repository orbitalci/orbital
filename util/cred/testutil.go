package cred

import (
	"bitbucket.org/level11consulting/go-til/consul"
	"bitbucket.org/level11consulting/go-til/vault"
	"bitbucket.org/level11consulting/ocelot/util"
	"errors"
	"fmt"
	"github.com/hashicorp/consul/testutil"
	"github.com/hashicorp/vault/http"
	hashiVault "github.com/hashicorp/vault/vault"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
)

//dummy vcscred obj so we don't have to depend on models anymore
type VcsConfig struct {
	ClientId     string
	ClientSecret string
	TokenURL     string
	AcctName     string
	Type         string
}

func (m *VcsConfig) GetAcctName() string {
	return m.AcctName
}

func (m *VcsConfig) GetType() string {
	return m.Type
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

func (m *RepoConfig) GetAcctName() string {
	return m.AcctName
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


func SetStoragePostgres(consulet *consul.Consulet, vaulty vault.Vaulty, dbName string, location string, port string, username string, pw string) (err error){
	err = consulet.AddKeyValue(StorageType, []byte("postgres"))
	if err != nil {
		return
	}
	err = consulet.AddKeyValue(PostgresDatabaseName, []byte(dbName))
	if err != nil {
		return
	}
	err = consulet.AddKeyValue(PostgresLocation, []byte(location))
	if err != nil {
		return
	}
	err = consulet.AddKeyValue(PostgresPort, []byte(port))
	if err != nil {
		return
	}
	err = consulet.AddKeyValue(PostgresUsername, []byte(username))
	if err != nil {
		return
	}
	var a = map[string]interface{} {PostgresPasswordKey: pw}
	_, err = vaulty.AddVaultData(PostgresPasswordLoc, a)
	return err
}


//////test setup and tear down///////

func TestSetupVaultAndConsul(t *testing.T) (CVRemoteConfig, net.Listener, *testutil.TestServer) {
	//set up unsealed vault for testing
	util.BuildServerHack(t)
	ln, token := TestSetupVault(t)

	//setup consul for testing
	testServer, host, port := TestSetupConsul(t)
	remoteConfig, err := GetInstance(host, port, token)
	if err != nil {
		t.Fatal(err)
	}
	return remoteConfig, ln, testServer
}


func TestSetupConsul(t *testing.T) (*testutil.TestServer, string, int) {
	//setup consul for testing
	testServer, err := testutil.NewTestServer()
	if err != nil {
		t.Fatal("Couldn't create consul test server, error: ", err)
	}
	ayy := strings.Split(testServer.HTTPAddr, ":")
	port, _ := strconv.ParseInt(ayy[1], 10, 32)
	return testServer, ayy[0], int(port)
}

func TestSetupVault(t *testing.T) (net.Listener, string) {
	core, _, token := hashiVault.TestCoreUnsealed(t)
	ln, addr := http.TestServer(t, core)
	os.Setenv("VAULT_ADDR", addr)
	os.Setenv("VAULT_TOKEN", token)
	return ln, token
}

func TeardownVaultAndConsul(testvault net.Listener, testconsul *testutil.TestServer) {
	testconsul.Stop()
	testvault.Close()
}
//
//func AddDockerRepoCreds(t *testing.T, rc CVRemoteConfig, repourl, password, username, acctName, projectName string) {
//	creds := &RepoConfig{
//		Password: password,
//		Username: username,
//		RepoUrl: map[string]string{"url":repourl},
//		Type: "docker",
//		AcctName: acctName,
//		ProjectName: projectName,
//	}
//	//err := testRemoteConfig.AddCreds(BuildCredPath("github", "mariannefeng", Vcs), adminConfig)
//	if err := rc.AddCreds(BuildCredPath("docker", acctName, Repo), creds); err != nil {
//		t.Fatal("couldnt add creds, error: ", err.Error())
//	}
//}
//
//func AddMvnRepoCreds(t *testing.T, rc CVRemoteConfig, repourl, password, username, acctName, projectName string) {
//	creds := &RepoConfig{
//		Password: password,
//		Username: username,
//		RepoUrl: map[string]string{"registry":repourl},
//		Type: "maven",
//		AcctName: acctName,
//		ProjectName: projectName,
//	}
//	if err := rc.AddCreds(BuildCredPath("maven", acctName, Repo), creds); err != nil {
//		t.Fatal("couldnt' add maven creds, error: ", err.Error())
//	}
//}

//type HealthyMaintainer interface {
//	Reconnect() error
//	Healthy() bool
//}

func NewHealthyMaintain() *HealthyMaintain {
	return &HealthyMaintain{true, true}
}
type HealthyMaintain struct {
	SuccessfulReconnect bool
	IsHealthy bool
}

func (h *HealthyMaintain) Reconnect() error {
	if h.SuccessfulReconnect {
		return nil
	}
	return errors.New("no reconnect for u bud")
}

func (h *HealthyMaintain) Healthy() bool {
	return h.IsHealthy
}
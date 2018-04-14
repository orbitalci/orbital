package credentials

import (
	"bitbucket.org/level11consulting/go-til/consul"
	"bitbucket.org/level11consulting/go-til/vault"
	"bitbucket.org/level11consulting/ocelot/old/admin/models"
	"bitbucket.org/level11consulting/ocelot/util"
	"errors"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"github.com/hashicorp/consul/testutil"
	"github.com/hashicorp/vault/http"
	hashiVault "github.com/hashicorp/vault/vault"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
)


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

func AddDockerRepoCreds(t *testing.T, rc CVRemoteConfig, store storage.CredTable, repourl, password, username, acctName, projectName string) {
	creds := &models.RepoCreds{
		Password: password,
		Username: username,
		RepoUrl: repourl,
		SubType: models.SubCredType_DOCKER,
		AcctName: acctName,
		Identifier: projectName,
	}
	if err := rc.AddCreds(store, creds, true); err != nil {
		t.Fatal("couldnt add creds, error: ", err.Error())
	}
}
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
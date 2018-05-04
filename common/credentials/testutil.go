package credentials

import (
	"github.com/hashicorp/consul/testutil"
	"github.com/hashicorp/vault/http"
	hashiVault "github.com/hashicorp/vault/vault"
	"github.com/shankj3/go-til/consul"
	"github.com/shankj3/go-til/vault"
	"github.com/shankj3/ocelot/common"
	tu "github.com/shankj3/ocelot/common/testutil"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"

	"errors"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
)

func SetStoragePostgres(consulet *consul.Consulet, vaulty vault.Vaulty, dbName string, location string, port string, username string, pw string) (err error) {
	err = consulet.AddKeyValue(common.StorageType, []byte("postgres"))
	if err != nil {
		return
	}
	err = consulet.AddKeyValue(common.PostgresDatabaseName, []byte(dbName))
	if err != nil {
		return
	}
	err = consulet.AddKeyValue(common.PostgresLocation, []byte(location))
	if err != nil {
		return
	}
	err = consulet.AddKeyValue(common.PostgresPort, []byte(port))
	if err != nil {
		return
	}
	err = consulet.AddKeyValue(common.PostgresUsername, []byte(username))
	if err != nil {
		return
	}
	var a = map[string]interface{}{common.PostgresPasswordKey: pw}
	_, err = vaulty.AddVaultData(common.PostgresPasswordLoc, a)
	return err
}

//////test setup and tear down///////

func TestSetupVaultAndConsul(t *testing.T) (CVRemoteConfig, net.Listener, *testutil.TestServer) {
	//set up unsealed vault for testing
	tu.BuildServerHack(t)
	ln, token := TestSetupVault(t)

	//setup consul for testing
	testServer, host, port := TestSetupConsul(t)
	remoteConfig, err := GetInstance(host, port, token)
	if err != nil {
		t.Error(err)
	}
	return remoteConfig, ln, testServer
}

func TestSetupConsul(t *testing.T) (*testutil.TestServer, string, int) {
	//setup consul for testing
	testServer, err := testutil.NewTestServer()
	if err != nil {
		t.Error("Couldn't create consul test server, error: ", err)
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
	creds := &pb.RepoCreds{
		Password:   password,
		Username:   username,
		RepoUrl:    repourl,
		SubType:    pb.SubCredType_DOCKER,
		AcctName:   acctName,
		Identifier: projectName,
	}
	if err := rc.AddCreds(store, creds, true); err != nil {
		t.Error("couldnt add creds, error: ", err.Error())
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
	IsHealthy           bool
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

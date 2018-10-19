package credentials

import (
	"sync"

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

// SetStoragePostgres Instantiates Consul and Vault instances with configuration values for postgresql server usage for static or dynamic credentials
// FIXME: vaulty is a borrowed value, so we should use as a reference
func SetStoragePostgres(consulet *consul.Consulet, vaulty vault.Vaulty, dbName string, location string, port string, username string, pw string, vaultEngine string, vaultRole string) (err error) {
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

	err = consulet.AddKeyValue(common.VaultDBSecretEngine, []byte(vaultEngine))
	if err != nil {
		return
	}

	err = consulet.AddKeyValue(common.VaultRoleName, []byte(vaultRole))
	if err != nil {
		return
	}

	switch vaultEngine {
	case "kv":
		// if the Vault DB Secret Engine is "kv", then we have to write the password in vault
		var a = map[string]interface{}{common.PostgresPasswordKey: pw}
		_, err = vaulty.AddVaultData(common.PostgresPasswordLoc, a)
		return err
	case "database":
		return errors.New("This test case is incomplete. This test requires go-til support for enabling the Database secret engine + creating a role")
	//	// if the Vault DB Secret Engine is "database", then we have to set up the database engine, and create a role
	//	var configPayload = map[string]interface{}{
	//		"plugin_name":             "postgresql-database-plugin",
	//		"allowed_roles":           fmt.Sprintf("%s", vaultRole),
	//		"connection_url":          fmt.Sprintf("postgresql://{{username}}:{{password}}@%s:%s/%s/?sslmode=disable", location, port, dbName),
	//		"max_open_connections":    5,
	//		"max_connection_lifetime": "5s",
	//		"username":                fmt.Sprintf("%s", username),
	//		"password":                fmt.Sprintf("%s", pw),
	//	}
	//	//_, err = vaulty.AddVaultData("/database/config/ocelot", configPayload)
	//	err = vaulty.EnableDatabaseSecretEngine(configPayload)
	//	if err != nil {
	//		return err
	//	}

	//	var rolePayload = map[string]interface{}{
	//		"db_name":             fmt.Sprintf("%s", dbName),
	//		"creation_statements": []string{"CREATE ROLE '{{name}}' WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; GRANT SELECT ON ALL TABLES IN SCHEMA public TO '{{name}}';"},
	//		"default_ttl":         "30s",
	//		"max_ttl":             "10m",
	//	}
	//	_, err = vaulty.AddVaultData(fmt.Sprintf("/database/role/%s", vaultRole), rolePayload)
	//	return err

	default:
		return errors.New("We only support 'kv' or 'database' secret engines in Vault")
	}

}

//////test setup and tear down///////

// TestSetupVaultAndConsul Returns an initialized remoteconfig, Vault, and Consul instances. Caller is responsible for calling TeardownVaultAndConsul() to close the connections
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

func NewHealthyMaintain() *HealthyMaintain {
	return &HealthyMaintain{SuccessfulReconnect: true, IsHealthy: true}
}

type HealthyMaintain struct {
	sync.Mutex
	SuccessfulReconnect bool
	IsHealthy           bool
}

func (h *HealthyMaintain) Reconnect() error {
	h.Lock()
	defer h.Unlock()
	if h.SuccessfulReconnect {
		return nil
	}
	return errors.New("no reconnect for u bud")
}

func (h *HealthyMaintain) Healthy() bool {
	h.Lock()
	defer h.Unlock()
	return h.IsHealthy
}

func (h *HealthyMaintain) SetSuccessfulReconnect() {
	h.Lock()
	defer h.Unlock()
	h.SuccessfulReconnect = true
}

func (h *HealthyMaintain) SetUnSuccessfulReconnect() {
	h.Lock()
	defer h.Unlock()
	h.SuccessfulReconnect = false
}

func (h *HealthyMaintain) SetUnHealthy() {
	h.Lock()
	defer h.Unlock()
	h.IsHealthy = false
}

func (h *HealthyMaintain) SetHealthy() {
	h.Lock()
	defer h.Unlock()
	h.IsHealthy = true
}

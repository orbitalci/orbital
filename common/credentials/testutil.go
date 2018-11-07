package credentials

import (
	"os/exec"
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
	"fmt"
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

		// If the Vault DB Secret Engine is "database", then we have to set up the database engine, and create a role

		// FIXME: We are expecting to use a different Vault instance (than the test vault) because the cli commands time out otherwise.
		// We use the vault cli, since the api is really involved with enabling backends, and creating roles. I'm lazy... ¯\_(ツ)_/¯
		vaultDbConfig := "vault write database/config/ocelot " +
			"plugin_name=postgresql-database-plugin " +
			"allowed_roles='ocelot' " +
			"connection_url='postgresql://{{username}}:{{password}}@192.168.56.78:5432/?sslmode=disable' " +
			"username='postgres' " +
			"password='mysecretpassword'"

		// FYI: This is a raw string bc of the nested quotes
		vaultDbRole := `vault write database/roles/ocelot \
        db_name=ocelot \
        creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
            GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
        default_ttl="10m" \
        max_ttl="1h"`

		fmt.Printf("Vault config! %s\n", vaultDbConfig)
		fmt.Printf("Vault role! %s\n", vaultDbRole)

		vagrantVaultScheme := "http"
		vagrantVaultIP := "192.168.56.78"
		vagrantVaultPort := 8200
		vagrantVaultToken := "ocelotdev"

		vaultAddrStr := fmt.Sprintf("VAULT_ADDR=%s", fmt.Sprintf("%s://%s:%d", vagrantVaultScheme, vagrantVaultIP, vagrantVaultPort))
		vaultTokenStr := fmt.Sprintf("VAULT_TOKEN=%s", vagrantVaultToken)

		// Turn on database secrets backend
		cmd := exec.Command("sh", "-c", fmt.Sprintf("vault secrets enable database || true"))
		cmd.Env = append(cmd.Env, vaultAddrStr)
		cmd.Env = append(cmd.Env, vaultTokenStr)

		_, err := cmd.CombinedOutput()
		if err != nil {
			return errors.New("Got an error trying to enable database secrets backend")
		}

		// Configure the Vault database secrets backend
		cmd = exec.Command("sh", "-c", fmt.Sprintf("%s", vaultDbConfig))
		cmd.Env = append(cmd.Env, vaultAddrStr)
		cmd.Env = append(cmd.Env, vaultTokenStr)

		_, err = cmd.CombinedOutput()
		if err != nil {
			return errors.New("Got an error trying to load Vault database backend config")
		}

		// Configure Ocelot's Vault role
		cmd = exec.Command("sh", "-c", fmt.Sprintf("%s", vaultDbRole))
		cmd.Env = append(cmd.Env, vaultAddrStr)
		cmd.Env = append(cmd.Env, vaultTokenStr)

		_, err = cmd.CombinedOutput()
		if err != nil {
			return errors.New("Got an error trying to load Vault ocelot role configuration")
		}

		return nil

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

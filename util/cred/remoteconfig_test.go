package cred

import (
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/vault"
	"os"
	"testing"
	"github.com/hashicorp/consul/testutil"
	"strings"
	"strconv"
	"net"
	"github.com/shankj3/ocelot/admin/models"
	"github.com/shankj3/ocelot/util/test"
)



func TestRemoteConfig_ErrorHandling(t *testing.T) {
	brokenRemote, _ := GetInstance("", 0, "abc")
	if brokenRemote == nil {
		t.Error(test.GenericStrFormatErrors("broken remote config", "not nil", brokenRemote))
	}
	err := brokenRemote.AddCreds("test", &models.Credentials{})
	if err.Error() != "not connected to consul, unable to add credentials" {
		t.Error(test.GenericStrFormatErrors("not connected to consul error message", "not connected to consul, unable to add credentials", err))
	}
	_, err = brokenRemote.GetCredAt("test", false)
	if err.Error() != "not connected to consul, unable to retrieve credentials" {
		t.Error(test.GenericStrFormatErrors("not connected to consul error message", "not connected to consul, unable to retrieve credentials", err))
	}

}

func TestRemoteConfig_OneGiantCredTest(t *testing.T) {
	testRemoteConfig, vaultListener, consulServer := testSetupVaultAndConsul(t)
	defer teardownVaultAndConsul(vaultListener, consulServer)

	adminConfig := &models.Credentials {
		ClientSecret: "top-secret",
		ClientId: "beeswax",
		AcctName: "mariannefeng",
		TokenURL: "a-real-url",
		Type: "github",
	}

	err := testRemoteConfig.AddCreds(ConfigPath + "/github/mariannefeng", adminConfig)
	if err != nil {
		t.Error(test.GenericStrFormatErrors("first adding creds to consul", nil, err))
	}

	testPassword, err := testRemoteConfig.GetPassword(ConfigPath + "/github/mariannefeng")
	if err != nil {
		t.Error(test.GenericStrFormatErrors("retrieving password", nil, err))
	}

	if testPassword != "top-secret" {
		t.Error(test.GenericStrFormatErrors("secret from vault", "top-secret", testPassword))
	}

	creds, _ := testRemoteConfig.GetCredAt(ConfigPath + "/github/mariannefeng", true)
	marianne, ok := creds["github/mariannefeng"]
	if !ok {
		t.Error(test.GenericStrFormatErrors("fake cred should exist", true, ok))
	}

	if marianne.AcctName != "mariannefeng" {
		t.Error(test.GenericStrFormatErrors("fake cred acct name", "mariannefeng", marianne.AcctName))
	}

	if marianne.TokenURL != "a-real-url" {
		t.Error(test.GenericStrFormatErrors("fake cred token url", "a-real-url", marianne.TokenURL))
	}

	if marianne.ClientId != "beeswax" {
		t.Error(test.GenericStrFormatErrors("fake cred client id", "beeswax", marianne.ClientId))
	}

	if marianne.Type != "github" {
		t.Error(test.GenericStrFormatErrors("fake cred acct type", "github", marianne.Type))
	}

	if marianne.ClientSecret != "*********" {
		t.Error(test.GenericStrFormatErrors("fake cred hidden password", "*********", marianne.ClientSecret))
	}

	creds, _ = testRemoteConfig.GetCredAt(ConfigPath + "/github/mariannefeng", false)
	marianne, _ = creds["github/mariannefeng"]

	if marianne.ClientSecret != "top-secret" {
		t.Error(test.GenericStrFormatErrors("fake cred should get password", "top-secret", marianne.ClientSecret))
	}

	secondConfig := &models.Credentials {
		ClientSecret: "secret",
		ClientId: "beeswaxxxxx",
		AcctName: "ariannefeng",
		TokenURL: "another-real-url",
		Type: "bitbucket",
	}

	err = testRemoteConfig.AddCreds(ConfigPath + "/bitbucket/ariannefeng", secondConfig)
	if err != nil {
		t.Error(test.GenericStrFormatErrors("adding second set of creds to consul", nil, err))
	}

	creds, _ = testRemoteConfig.GetCredAt(ConfigPath, false)

	_, ok = creds["github/mariannefeng"]
	if !ok {
		t.Error(test.GenericStrFormatErrors("original creds marianne should exist", true, ok))
	}
	newCreds, ok := creds["bitbucket/ariannefeng"]
	if !ok {
		t.Error(test.GenericStrFormatErrors("new creds arianne should exist", true, ok))
	}

	if newCreds.AcctName != "ariannefeng" {
		t.Error(test.GenericStrFormatErrors("2nd fake cred acct name", "ariannefeng", newCreds.AcctName))
	}

	if newCreds.TokenURL != "another-real-url" {
		t.Error(test.GenericStrFormatErrors("2nd fake cred token url", "another-real-url", newCreds.TokenURL))
	}

	if newCreds.ClientId != "beeswaxxxxx" {
		t.Error(test.GenericStrFormatErrors("2nd fake cred client id", "beeswaxxxxx", newCreds.ClientId))
	}

	if newCreds.Type != "bitbucket" {
		t.Error(test.GenericStrFormatErrors("2nd fake cred acct type", "bitbucket", newCreds.Type))
	}

	if newCreds.ClientSecret != "secret" {
		t.Error(test.GenericStrFormatErrors("2nd fake open password", "secret", newCreds.ClientSecret))
	}

}


//////test setup and tear down///////

func testSetupVaultAndConsul(t *testing.T) (*RemoteConfig, net.Listener, *testutil.TestServer) {
	//set up unsealed vault for testing
	core, _, token := vault.TestCoreUnsealed(t)
	ln, addr := http.TestServer(t, core)
	os.Setenv("VAULT_ADDR", addr)
	os.Setenv("VAULT_TOKEN", token)

	//setup consul for testing
	testServer, err := testutil.NewTestServer()
	if err != nil {
		t.Fatal("Couldn't create consul test server, error: ", err)
	}
	ayy := strings.Split(testServer.HTTPAddr, ":")
	port, _ := strconv.ParseInt(ayy[1], 10, 32)

	remoteConfig, err := GetInstance(ayy[0], int(port), token)

	return remoteConfig, ln, testServer
}

func teardownVaultAndConsul(testvault net.Listener, testconsul *testutil.TestServer) {
	testconsul.Stop()
	testvault.Close()
}
package cred

import (
	"bitbucket.org/level11consulting/go-til/test"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"github.com/hashicorp/consul/testutil"
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/vault"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestRemoteConfig_ErrorHandling(t *testing.T) {
	brokenRemote, _ := GetInstance("", 19000, "abc")
	if brokenRemote == nil {
		t.Error(test.GenericStrFormatErrors("broken remote config", "not nil", brokenRemote))
	}
	err := brokenRemote.AddCreds("test", &models.Credentials{})
	if err.Error() != "not connected to consul, unable to add credentials" {
		t.Error(test.GenericStrFormatErrors("not connected to consul error message", "not connected to consul, unable to add credentials", err))
	}
	_, err = brokenRemote.GetCredAt("test", false, Vcs)
	if err.Error() != "not connected to consul, unable to retrieve credentials" {
		t.Error(test.GenericStrFormatErrors("not connected to consul error message", "not connected to consul, unable to retrieve credentials", err))
	}

}

func TestRemoteConfig_OneGiantCredTest(t *testing.T) {
	testRemoteConfig, vaultListener, consulServer := testSetupVaultAndConsul(t)
	defer teardownVaultAndConsul(vaultListener, consulServer)

	adminConfig := &models.Credentials{
		ClientSecret: "top-secret",
		ClientId:     "beeswax",
		AcctName:     "mariannefeng",
		TokenURL:     "a-real-url",
		Type:         "github",
	}
	err := testRemoteConfig.AddCreds(BuildCredPath("github", "mariannefeng", Vcs), adminConfig)
	if err != nil {
		t.Error(test.GenericStrFormatErrors("first adding creds to consul", nil, err))
	}

	testPassword, err := testRemoteConfig.GetPassword(BuildCredPath("github", "mariannefeng", Vcs))
	if err != nil {
		t.Error(test.GenericStrFormatErrors("retrieving password", nil, err))
	}

	if testPassword != "top-secret" {
		t.Error(test.GenericStrFormatErrors("secret from vault", "top-secret", testPassword))
	}

	creds, _ := testRemoteConfig.GetCredAt(BuildCredPath("github", "mariannefeng", Vcs), true, Vcs)
	mari, ok := creds["github/mariannefeng"]
	if !ok {
		t.Error(test.GenericStrFormatErrors("fake cred should exist", true, ok))
	}
	marianne, ok := mari.(*models.Credentials)
	if !ok {
		t.Fatal("could not cast GetCredAt cred.RemoteConfigCred interface to adminConfig models.Credentials")
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

	creds, _ = testRemoteConfig.GetCredAt(BuildCredPath("github", "mariannefeng", Vcs), false, Vcs)
	mari, _ = creds["github/mariannefeng"]
	marianne, ok = mari.(*models.Credentials)
	if !ok {
		t.Fatal("could not cast GetCredAt cred.RemoteConfigCred interface to adminConfig models.Credentials")
	}

	if marianne.ClientSecret != "top-secret" {
		t.Error(test.GenericStrFormatErrors("fake cred should get password", "top-secret", marianne.ClientSecret))
	}

	secondConfig := &models.Credentials{
		ClientSecret: "secret",
		ClientId:     "beeswaxxxxx",
		AcctName:     "ariannefeng",
		TokenURL:     "another-real-url",
		Type:         "bitbucket",
	}

	err = testRemoteConfig.AddCreds(BuildCredPath("bitbucket", "ariannefeng", Vcs), secondConfig)
	if err != nil {
		t.Error(test.GenericStrFormatErrors("adding second set of creds to consul", nil, err))
	}

	creds, _ = testRemoteConfig.GetCredAt(ConfigPath, false, Vcs)

	_, ok = creds["github/mariannefeng"]
	if !ok {
		t.Error(test.GenericStrFormatErrors("original creds marianne should exist", true, ok))
	}
	newCred, ok := creds["bitbucket/ariannefeng"]
	if !ok {
		t.Fatal(test.GenericStrFormatErrors("new creds arianne should exist", true, ok))
	}
	newCreds, ok := newCred.(*models.Credentials)
	if !ok {
		t.Fatal("could not cast GetCredAt cred.RemoteConfigCred interface to adminConfig models.Credentials")
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

	repoCreds := &models.RepoCreds{
		Username: "tasty-gummy-vitamin",
		Password: "FLINTSTONE",
		RepoUrl: "http://take-ur-vitamins.org/uploadGummy",
		AcctName: "jessdanshnak",
		Type: "nexus",
	}
	repoPath := BuildCredPath("nexus", "jessdanshnak", Repo)
	err = testRemoteConfig.AddCreds(repoPath, repoCreds)
	if err != nil {
		t.Error(test.GenericStrFormatErrors("adding repo creds", nil, err))
	}
	repoData, err := testRemoteConfig.GetCredAt(repoPath, false, Repo)
	if err != nil {
		t.Error(test.GenericStrFormatErrors("getting repo creds", nil , err))
	}
	shank, ok := repoData["nexus/jessdanshnak"]
	if !ok {
		t.Fatal("inserted repo creds w/ path nexus/jessdanshnak should exist")
	}
	shnak, ok := shank.(*models.RepoCreds)
	if !ok {
		t.Fatal("could not cast GetCredAt cred.RemoteConfigCred interface to repo config *models.RepoCreds")
	}
	if shnak.GetPassword() != repoCreds.GetPassword() {
		t.Error(test.StrFormatErrors("repo password", repoCreds.Password, shnak.Password))
	}
	if shnak.GetType() != repoCreds.GetType() {
		t.Error(test.StrFormatErrors("repo acct type", repoCreds.GetType(), shnak.GetType()))
	}
	if shnak.GetRepoUrl() != repoCreds.GetRepoUrl() {
		t.Error(test.StrFormatErrors("repo url", repoCreds.GetRepoUrl(), shnak.GetRepoUrl()))
	}
	// testing that all creds should still be there
	creds, _ = testRemoteConfig.GetCredAt(ConfigPath, false, Vcs)
	if _, ok = creds["bitbucket/ariannefeng"]; !ok {
		t.Error("there should still be the admin credentials at bitbucket/ariannefeng")
	}
	if _, ok = creds["github/mariannefeng"]; !ok {
		t.Error("there should still be admin credentials at github/mariannefeng")
	}
}


func Test_BuildCredPath(t *testing.T) {
	expected := "creds/vcs/banana/bitbucket"
	live := BuildCredPath("bitbucket", "banana", Vcs)
	if live != expected {
		t.Error(test.StrFormatErrors("vcs cred path", expected, live))
	}
	expectedRepo := "creds/repo/jessjess/nexus"
	liveRepo := BuildCredPath("nexus", "jessjess", Repo)
	if liveRepo != expectedRepo {
		t.Error(test.StrFormatErrors("repo cred path", expectedRepo, liveRepo))
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

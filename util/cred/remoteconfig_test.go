package cred

import (
	"bitbucket.org/level11consulting/go-til/consul"
	"bitbucket.org/level11consulting/go-til/test"
	"bitbucket.org/level11consulting/go-til/vault"
	"bitbucket.org/level11consulting/ocelot/util"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"testing"
	"time"
)

//func TestRemoteConfig_ErrorHandling(t *testing.T) {
//	brokenRemote, _ := GetInstance("localhost", 19000, "abc")
//	if brokenRemote == nil {
//		t.Error(test.GenericStrFormatErrors("broken remote config", "not nil", brokenRemote))
//	}
//	err := brokenRemote.AddCreds("test", &VcsConfig{})
//	if err.Error() != "not connected to consul, unable to add credentials" {
//		t.Error(test.GenericStrFormatErrors("not connected to consul error message", "not connected to consul, unable to add credentials", err))
//	}
//	_, err = brokenRemote.GetCredAt("test", false, &VcsConfig{})
//	if err.Error() != "not connected to consul, unable to retrieve credentials" {
//		t.Error(test.GenericStrFormatErrors("not connected to consul error message", "not connected to consul, unable to retrieve credentials", err))
//	}
//
//}

func TestRemoteConfig_OneGiantCredTest(t *testing.T) {

	testRemoteConfig, vaultListener, consulServer := TestSetupVaultAndConsul(t)
	defer TeardownVaultAndConsul(vaultListener, consulServer)
	adminConfig := &VcsConfig{
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

	creds, _ := testRemoteConfig.GetCredAt(BuildCredPath("github", "mariannefeng", Vcs), true, &VcsConfig{})
	mari, ok := creds["github/mariannefeng"]
	if !ok {
		t.Error(test.GenericStrFormatErrors("fake cred should exist", true, ok))
	}
	marianne, ok := mari.(*VcsConfig)
	if !ok {
		t.Fatal("could not cast GetCredAt cred.RemoteConfigCred interface to adminConfig models.VCSCreds")
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

	creds, _ = testRemoteConfig.GetCredAt(BuildCredPath("github", "mariannefeng", Vcs), false, &VcsConfig{})
	mari, _ = creds["github/mariannefeng"]
	marianne, ok = mari.(*VcsConfig)
	if !ok {
		t.Fatal("could not cast GetCredAt cred.RemoteConfigCred interface to adminConfig models.VCSCreds")
	}

	if marianne.ClientSecret != "top-secret" {
		t.Error(test.GenericStrFormatErrors("fake cred should get password", "top-secret", marianne.ClientSecret))
	}

	secondConfig := &VcsConfig{
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

	creds, _ = testRemoteConfig.GetCredAt(ConfigPath, false, &VcsConfig{})

	_, ok = creds["github/mariannefeng"]
	if !ok {
		t.Error(test.GenericStrFormatErrors("original creds marianne should exist", true, ok))
	}
	newCred, ok := creds["bitbucket/ariannefeng"]
	if !ok {
		t.Fatal(test.GenericStrFormatErrors("new creds arianne should exist", true, ok))
	}
	newCreds, ok := newCred.(*VcsConfig)
	if !ok {
		t.Fatal("could not cast GetCredAt cred.RemoteConfigCred interface to adminConfig models.VCSCreds")
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

	repoCreds := &RepoConfig{
		Username: "tasty-gummy-vitamin",
		Password: "FLINTSTONE",
		RepoUrl: map[string]string{"":"http://take-ur-vitamins.org/uploadGummy"},
		AcctName: "jessdanshnak",
		Type: "nexus",
	}
	repoPath := BuildCredPath("nexus", "jessdanshnak", Repo)
	err = testRemoteConfig.AddCreds(repoPath, repoCreds)
	if err != nil {
		t.Error(test.GenericStrFormatErrors("adding repo creds", nil, err))
	}
	repoData, err := testRemoteConfig.GetCredAt(repoPath, false, &RepoConfig{})
	if err != nil {
		t.Error(test.GenericStrFormatErrors("getting repo creds", nil , err))
	}
	shank, ok := repoData["nexus/jessdanshnak"]
	if !ok {
		t.Fatal("inserted repo creds w/ path nexus/jessdanshnak should exist")
	}
	shnak, ok := shank.(*RepoConfig)
	if !ok {
		t.Fatal("could not cast GetCredAt cred.RemoteConfigCred interface to repo config *models.RepoCreds")
	}
	if shnak.GetPassword() != repoCreds.GetPassword() {
		t.Error(test.StrFormatErrors("repo password", repoCreds.Password, shnak.Password))
	}
	if shnak.GetType() != repoCreds.GetType() {
		t.Error(test.StrFormatErrors("repo acct type", repoCreds.GetType(), shnak.GetType()))
	}
	//if shnak.GetRepoUrl() != repoCreds.GetRepoUrl() {
	//	t.Error(test.StrFormatErrors("repo url", repoCreds.GetRepoUrl(), shnak.GetRepoUrl()))
	//}
	// testing that all creds should still be there
	creds, _ = testRemoteConfig.GetCredAt(ConfigPath, false, &VcsConfig{})
	if _, ok = creds["bitbucket/ariannefeng"]; !ok {
		t.Error("there should still be the admin credentials at bitbucket/ariannefeng")
	}
	if _, ok = creds["github/mariannefeng"]; !ok {
		t.Error("there should still be admin credentials at github/mariannefeng")
	}
}

func TestRemoteConfig_GetStorageType(t *testing.T) {
	util.BuildServerHack(t)
	testRemoteConfig, vaultListener, consulServer := TestSetupVaultAndConsul(t)
	defer TeardownVaultAndConsul(vaultListener, consulServer)
	// check that default will be file
	storeType, err := testRemoteConfig.GetStorageType()
	if err != nil {
		t.Fatal("should be able to get storage type, err: " + err.Error())
	}
	if storeType != storage.FileSystem {
		t.Error(test.GenericStrFormatErrors("type", storage.FileSystem, storeType))
	}
	consulServer.SetKV(t, "config/ocelot/storagetype", []byte("filesystem"))
	storeType, err = testRemoteConfig.GetStorageType()
	if err != nil {
		t.Fatal("should be able to get storage type, err: " + err.Error())
	}
	if storeType != storage.FileSystem {
		t.Error(test.GenericStrFormatErrors("store type enum", storage.FileSystem, storeType))
	}
	consulServer.SetKV(t, "config/ocelot/storagetype", []byte("postgres"))
	storeType, err = testRemoteConfig.GetStorageType()
	if err != nil {
		t.Fatal("should be able to get storage type, err: " + err.Error())
	}
	if storeType != storage.Postgres {
		t.Error(test.GenericStrFormatErrors("store type enum", storage.Postgres, storeType))
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

func TestRemoteConfig_Healthy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestRemoteConfig_Healthy because requires killing / restarting vault and consul multiple times.")
	}
	testRemoteConfig, vaultListener, consulServer := TestSetupVaultAndConsul(t)
	if !testRemoteConfig.Healthy() {
		t.Error("consul & vault are up, should return status of healthy")
	}
	consulServer.Stop()
	time.Sleep(1*time.Second)
	if testRemoteConfig.Healthy() {
		t.Error("consul has been taken down, should return status of not healthy.")
	}
	newConsul, host, port := TestSetupConsul(t)
	defer newConsul.Stop()
	rc := testRemoteConfig.(*RemoteConfig)
	var err error
	rc.Consul, err = consul.New(host, port)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1*time.Second)
	if !testRemoteConfig.Healthy() {
		t.Error("consul has been stood back up, should return status of healthy.")
	}
	vaultListener.Close()
	time.Sleep(1*time.Second)
	if testRemoteConfig.Healthy() {
		t.Error("vault has been shut down, shouldn ot return status of healthy.")
	}
}

func TestRemoteConfig_Reconnect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestRemoteConfig_Reconnect because requires killing / restarting vault and consul multiple times.")
	}
	testRemoteConfig, vaultListener, consulServer := TestSetupVaultAndConsul(t)
	rc := testRemoteConfig.(*RemoteConfig)
	if err := testRemoteConfig.Reconnect(); err != nil {
		t.Error("should be able to 'reconnect' as both vault and consul are up, instead error is ", err.Error())
	}
	vaultListener.Close()
	//vaultClient, err := ocevault.NewAuthedClient(token)
	time.Sleep(1*time.Second)
	if err := testRemoteConfig.Reconnect(); err == nil {
		t.Error("should not be able to 'reconnect' as vault is down.")
	}
	newVaultListener, token := TestSetupVault(t)
	defer newVaultListener.Close()
	vaultClient, err := vault.NewAuthedClient(token)
	if err != nil {
		t.Error(err)
	}
	rc.Vault = vaultClient
	time.Sleep(500*time.Millisecond)
	if err := testRemoteConfig.Reconnect(); err != nil {
		t.Error("should be able to 'reconnect' because vault has been stood back up. instead the error is ", err.Error())
	}
	consulServer.Stop()
	time.Sleep(3*time.Second)
	if err := testRemoteConfig.Reconnect(); err == nil {
		t.Error("should not be able to 'reconnect' as consul is down.")
	}
	newConsul, host, port := TestSetupConsul(t)
	defer newConsul.Stop()
	rc.Consul, err = consul.New(host, port)
	if err != nil {
		t.Error(err)
	}
	time.Sleep(2*time.Second)
	if err := testRemoteConfig.Reconnect(); err != nil {
		t.Error("should be able to 'reconnect' because consul has been stood back up. instead the error is ", err.Error())
	}

}
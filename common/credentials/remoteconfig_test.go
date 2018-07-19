package credentials

import (
	"github.com/shankj3/go-til/consul"
	"github.com/shankj3/go-til/test"
	"github.com/shankj3/go-til/vault"
	pb "github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"

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
	port := 5439
	cleanup, pw := storage.CreateTestPgDatabase(t, port)
	defer cleanup(t)
	pg := storage.NewPostgresStorage("postgres", pw, "localhost", port, "postgres")
	adminConfig := &pb.VCSCreds{
		ClientSecret: "top-secret",
		ClientId:     "beeswax",
		AcctName:     "mariannefeng",
		TokenURL:     "a-real-url",
		Identifier:   "123",
		SubType:      pb.SubCredType_GITHUB,
	}
	err := testRemoteConfig.AddCreds(pg, adminConfig, true)
	if err != nil {
		t.Error(test.GenericStrFormatErrors("first adding creds to consul", nil, err))
	}

	testPassword, err := testRemoteConfig.GetPassword(pb.SubCredType_GITHUB, adminConfig.AcctName, pb.CredType_VCS, "123")
	if err != nil {
		t.Error(test.GenericStrFormatErrors("retrieving password", nil, err))
	}

	if testPassword != "top-secret" {
		t.Error(test.GenericStrFormatErrors("secret from vault", "top-secret", testPassword))
	}

	mari, err := testRemoteConfig.GetCred(pg, pb.SubCredType_GITHUB, "123", "mariannefeng", true)
	if err != nil {
		t.Fatal(err)
	}
	marianne, ok := mari.(*pb.VCSCreds)
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

	if marianne.SubType != pb.SubCredType_GITHUB {
		t.Error(test.GenericStrFormatErrors("fake cred acct type", "github", marianne.SubType))
	}

	if marianne.ClientSecret != "*********" {
		t.Error(test.GenericStrFormatErrors("fake cred hidden password", "*********", marianne.ClientSecret))
	}

	mari, err = testRemoteConfig.GetCred(pg, pb.SubCredType_GITHUB, "123", "mariannefeng", false)
	if err != nil {
		t.Fatal(err)
	}
	marianne, ok = mari.(*pb.VCSCreds)
	if !ok {
		t.Fatal("could not cast GetCredAt cred.RemoteConfigCred interface to adminConfig models.VCSCreds")
	}

	if marianne.ClientSecret != "top-secret" {
		t.Error(test.GenericStrFormatErrors("fake cred should get password", "top-secret", marianne.ClientSecret))
	}

	secondConfig := &pb.VCSCreds{
		ClientSecret: "secret",
		ClientId:     "beeswaxxxxx",
		AcctName:     "ariannefeng",
		TokenURL:     "another-real-url",
		Identifier:   "345",
		SubType:      pb.SubCredType_BITBUCKET,
	}

	err = testRemoteConfig.AddCreds(pg, secondConfig, true)
	if err != nil {
		t.Error(test.GenericStrFormatErrors("adding second set of creds to consul", nil, err))
	}

	cred, err := testRemoteConfig.GetCred(pg, pb.SubCredType_GITHUB, "123", "mariannefeng", false)
	if err != nil || cred == nil {
		t.Error(test.GenericStrFormatErrors("original creds marianne should exist", true, ok))
	}
	newCred, err := testRemoteConfig.GetCred(pg, pb.SubCredType_BITBUCKET, "345", "ariannefeng", false)
	if err != nil {
		t.Fatal(err)
	}
	newCreds, ok := newCred.(*pb.VCSCreds)
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

	if newCreds.SubType != pb.SubCredType_BITBUCKET {
		t.Error(test.GenericStrFormatErrors("2nd fake cred acct sub type", "bitbucket", newCreds.SubType))
	}

	if newCreds.ClientSecret != "secret" {
		t.Error(test.GenericStrFormatErrors("2nd fake open password", "secret", newCreds.ClientSecret))
	}

	repoCreds := &pb.RepoCreds{
		Username:   "tasty-gummy-vitamin",
		Password:   "FLINTSTONE",
		RepoUrl:    "http://take-ur-vitamins.org/uploadGummy",
		Identifier: "890",
		AcctName:   "jessdanshnak",
		SubType:    pb.SubCredType_NEXUS,
	}

	err = testRemoteConfig.AddCreds(pg, repoCreds, true)
	if err != nil {
		t.Error(test.GenericStrFormatErrors("adding repo creds", nil, err))
	}
	shank, err := testRemoteConfig.GetCred(pg, pb.SubCredType_NEXUS, "890", "jessdanshnak", false)
	if err != nil {
		t.Fatal(err)
	}
	shnak, ok := shank.(*pb.RepoCreds)
	if !ok {
		t.Fatal("could not cast GetCredAt cred.RemoteConfigCred interface to repo config *models.RepoCreds")
	}
	if shnak.GetPassword() != repoCreds.GetPassword() {
		t.Error(test.StrFormatErrors("repo password", repoCreds.Password, shnak.Password))
	}
	if shnak.GetRepoUrl() != repoCreds.GetRepoUrl() {
		t.Error(test.StrFormatErrors("repo url", repoCreds.GetRepoUrl(), shnak.GetRepoUrl()))
	}
	// test storage type
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
	expected := "creds/vcs/banana/bitbucket/derp"
	live := BuildCredPath(pb.SubCredType_BITBUCKET, "banana", pb.CredType_VCS, "derp")
	if live != expected {
		t.Error(test.StrFormatErrors("vcs cred path", expected, live))
	}
	expectedRepo := "creds/repo/jessjess/nexus/id123"
	liveRepo := BuildCredPath(pb.SubCredType_NEXUS, "jessjess", pb.CredType_REPO, "id123")
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
	time.Sleep(1 * time.Second)
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
	time.Sleep(1 * time.Second)
	if !testRemoteConfig.Healthy() {
		t.Error("consul has been stood back up, should return status of healthy.")
	}
	vaultListener.Close()
	time.Sleep(1 * time.Second)
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
	time.Sleep(1 * time.Second)
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
	time.Sleep(500 * time.Millisecond)
	if err := testRemoteConfig.Reconnect(); err != nil {
		t.Error("should be able to 'reconnect' because vault has been stood back up. instead the error is ", err.Error())
	}
	consulServer.Stop()
	time.Sleep(3 * time.Second)
	if err := testRemoteConfig.Reconnect(); err == nil {
		t.Error("should not be able to 'reconnect' as consul is down.")
	}
	newConsul, host, port := TestSetupConsul(t)
	defer newConsul.Stop()
	rc.Consul, err = consul.New(host, port)
	if err != nil {
		t.Error(err)
	}
	time.Sleep(2 * time.Second)
	if err := testRemoteConfig.Reconnect(); err != nil {
		t.Error("should be able to 'reconnect' because consul has been stood back up. instead the error is ", err.Error())
	}

}

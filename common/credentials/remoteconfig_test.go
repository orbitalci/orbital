package credentials

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/shankj3/go-til/consul"
	"github.com/shankj3/go-til/test"
	"github.com/shankj3/go-til/vault"
	"github.com/shankj3/ocelot/common"
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
	secondConfig.TokenURL = "bannananahammooooock"
	err = testRemoteConfig.UpdateCreds(pg, secondConfig)
	if err != nil {
		t.Error(err)
	}

	cred, err = testRemoteConfig.GetCred(pg, secondConfig.SubType, secondConfig.Identifier, secondConfig.AcctName, true)
	if err != nil {
		t.Error(err)
	}
	crreddy := cred.(*pb.VCSCreds)
	if crreddy.TokenURL != "bannananahammooooock" {
		t.Error(test.StrFormatErrors("identifier", "bannananahammooooock", crreddy.TokenURL))
	}

}

func TestRemoteConfig_BuildCredPath(t *testing.T) {
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

func TestRemoteConfig_DeleteCred(t *testing.T) {
	// happy path
	testCred := &pb.RepoCreds{SubType: pb.SubCredType_DOCKER, Identifier: "shazam", AcctName: "jazzy"}
	ctl := gomock.NewController(t)
	store := storage.NewMockOcelotStorage(ctl)
	safe := vault.NewMockVaulty(ctl)
	rc := &RemoteConfig{Vault: safe}
	store.EXPECT().DeleteCred(testCred).Return(nil).Times(1)
	safe.EXPECT().DeletePath("creds/repo/jazzy/docker/shazam").Return(nil).Times(1)
	err := rc.DeleteCred(store, testCred)
	if err != nil {
		t.Error(err)
	}
	// store fail
	store.EXPECT().DeleteCred(testCred).Return(storage.CredNotFound("jazzy", "docker")).Times(1)
	safe.EXPECT().DeletePath("creds/repo/jazzy/docker/shazam").Return(nil).Times(1)
	err = rc.DeleteCred(store, testCred)
	if err == nil {
		t.Error("an error should be returned as the store failed")
		return
	}
	if err.Error() != "unable to delete un-sensitive data: no credential found for jazzy docker" {
		t.Error("did not get expected error, got: " + err.Error())
	}
	// vault fail
	safe.EXPECT().DeletePath("creds/repo/jazzy/docker/shazam").Return(errors.New("oh jesus wtf happened ")).Times(1)
	store.EXPECT().DeleteCred(testCred).Return(nil).Times(1)
	err = rc.DeleteCred(store, testCred)
	if err == nil {
		t.Error("an error should be returned as the vault instance 'failed'")
		return
	}
	if err.Error() != "unable to delete sensitive data : Unable to delete password for user jazzy w/ identifier shazam: oh jesus wtf happened " {
		t.Error("did not get expected error, got: " + err.Error())
	}

	// vault & store fail
	safe.EXPECT().DeletePath("creds/repo/jazzy/docker/shazam").Return(errors.New("oh jesus wtf happened ")).Times(1)
	store.EXPECT().DeleteCred(testCred).Return(storage.CredNotFound("jazzy", "docker")).Times(1)
	err = rc.DeleteCred(store, testCred)
	if err.Error() != "unable to delete sensitive data : Unable to delete password for user jazzy w/ identifier shazam: oh jesus wtf happened : unable to delete un-sensitive data: no credential found for jazzy docker" {
		t.Error("did not get expected errors, got " + err.Error())
	}

}

func TestRemoteConfig_BuildCredKey(t *testing.T) {

	expectedCredKey := "credType/acctName"
	returnedCredKey := BuildCredKey("credType", "acctName")

	if returnedCredKey != expectedCredKey {
		t.Error("expected: "+returnedCredKey+"got :", expectedCredKey)
	}
}

func TestRemoteConfig_GetSetConsulVault(t *testing.T) {

	// GetInstance, error path
	consulHostNull := ""
	consulPortNull := 0
	vaultTokenNull := ""

	badRC, baderr := GetInstance(consulHostNull, consulPortNull, vaultTokenNull)

	expectedRemoteConfigString := "*credentials.RemoteConfig"
	remoteConfigString := fmt.Sprintf("%T", badRC)
	if remoteConfigString != expectedRemoteConfigString || baderr != nil {
		t.Error("Bad consulHost and port provided. Expected " + expectedRemoteConfigString + "and no errors, but got: " + fmt.Sprintf("%T", badRC) + "and " + fmt.Sprintf("%s", baderr))
	}

	// GetInstance, happy path, via TestSetupVaultAndConsul
	goodRemoteConfig, vaultListener, consulServer := TestSetupVaultAndConsul(t)
	defer TeardownVaultAndConsul(vaultListener, consulServer)

	// Verify Vault and Consul get/set integriy
	testVaultHandler := goodRemoteConfig.GetVault()
	goodRemoteConfig.SetVault(testVaultHandler)
	if testVaultHandler != goodRemoteConfig.GetVault() {
		t.Error("Vault get/set integrity fail")
	}

	testConsulHandler := goodRemoteConfig.GetConsul()
	goodRemoteConfig.SetConsul(testConsulHandler)
	if testConsulHandler != goodRemoteConfig.GetConsul() {
		t.Error("Consul get/set integrity fail")
	}

}

func TestRemoteConfig_GetToken(t *testing.T) {
	oldToken := os.Getenv("VAULT_TOKEN")

	if oldToken != "" {
		setupErr := os.Unsetenv("VAULT_TOKEN")
		if setupErr != nil {
			t.Error("VAULT_TOKEN should be unset now for this test, but it isn't...")
		}
	}

	badToken, badErr := GetToken("")

	if badToken != "" || badErr == nil {
		t.Error("VAULT_TOKEN env var isn't set. This token should be empty")
	}

	// FIXME: This case needs a way to force an I/O error when reading a real path. /dev/zero,/dev/null doesn't work
	// 			Try creating a temp file with no read permissions
	// os.Setenv("VAULT_TOKEN", "")
	// badToken, badErr = GetToken("/dev/null")

	// if badToken != "" || badErr == nil {
	// 	t.Error("VAULT_TOKEN should be empty, and a real filepath w/o a token")
	// }

	somethingToken, somethingErr := GetToken("/usr/share/dict/words")
	if somethingToken == "" || somethingErr != nil {
		t.Error("Assuming /usr/share/dict/words exists, this token should be set to a real value")
	}
}

func TestRemoteConfig_GetStorageType(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping docker container create due to -short being set")
	}

	testRemoteConfig, vaultListener, consulServer := TestSetupVaultAndConsul(t)
	defer TeardownVaultAndConsul(vaultListener, consulServer)

	// Consul StorageType unconfigured
	storageType, err := testRemoteConfig.GetStorageType()
	if err != nil {
		t.Error("This call should have failed, because Consul not fully initialized")
	}

	// StorageType is an invalid option
	err = testRemoteConfig.GetConsul().AddKeyValue(common.StorageType, []byte("punchcard"))
	if err != nil {
		return
	}
	storageType, err = testRemoteConfig.GetStorageType()

	if storageType != 0 || err == nil {
		t.Error("punchcard isn't a valid option. This should have failed")
	}

	// filesystem case
	err = testRemoteConfig.GetConsul().AddKeyValue(common.StorageType, []byte("filesystem"))
	if err != nil {
		return
	}

	storageType, err = testRemoteConfig.GetStorageType()

	switch storageType {
	case storage.Postgres:
		t.Error("The expected storageType is filesystem")
	case storage.FileSystem:
		fmt.Println("filesystem")
	default:
		t.Error("This should be an unreachable case - filesystem case")
	}

	// postgres case
	err = testRemoteConfig.GetConsul().AddKeyValue(common.StorageType, []byte("postgres"))
	if err != nil {
		return
	}

	storageType, err = testRemoteConfig.GetStorageType()

	switch storageType {
	case storage.Postgres:
		fmt.Println("postgres")
	case storage.FileSystem:
		t.Error("The expected storageType is postgres")
	default:
		t.Error("This should be an unreachable case - postgres case")
	}

}

func TestRemoteConfig_GetStorageCreds(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping docker container create due to -short being set")
	}

	testRemoteConfig, vaultListener, consulServer := TestSetupVaultAndConsul(t)
	defer TeardownVaultAndConsul(vaultListener, consulServer)

	// This case fails to due to incomplete implementation of filesystem StorageType
	//	// filesystem case
	//	err := testRemoteConfig.GetConsul().AddKeyValue(common.StorageType, []byte("filesystem"))
	//	if err != nil {
	//		return
	//	}
	//
	//	storageType, err := testRemoteConfig.GetStorageType()
	//
	//	switch storageType {
	//	case storage.Postgres:
	//		t.Error("The expected storageType is filesystem")
	//	case storage.FileSystem:
	//		fmt.Println("filesystem")
	//
	//		creds, err := testRemoteConfig.GetStorageCreds(&storageType)
	//		if err != nil {
	//			t.Error("Failed to get filesystem storageCreds")
	//		}
	//
	//		fmt.Println(creds)
	//	default:
	//		t.Error("This should be an unreachable case - filesystem case")
	//	}

	// garbage case
	err := testRemoteConfig.GetConsul().AddKeyValue(common.StorageType, []byte("garbage"))
	_, err = testRemoteConfig.GetStorageType()
	if err == nil {
		t.Error("This case is supposed to fail. There is no garbage storageType")
	}

	// postgres case
	err = testRemoteConfig.GetConsul().AddKeyValue(common.StorageType, []byte("postgres"))
	if err != nil {
		return
	}

	storageType, err := testRemoteConfig.GetStorageType()
	switch storageType {
	// Testing getForPostgres()
	case storage.Postgres:

		testRemoteConfig, vaultListener, consulServer := TestSetupVaultAndConsul(t)
		defer TeardownVaultAndConsul(vaultListener, consulServer)
		port := 5438
		cleanup, pw := storage.CreateTestPgDatabase(t, port)
		defer cleanup(t)
		//_ = storage.NewPostgresStorage("postgres", pw, "localhost", port, "postgres")

		// We are going to want to set these fields individually, to hit the error branches

		creds, err := testRemoteConfig.GetStorageCreds(&storageType)
		if err == nil {
			t.Error("This call should fail. Consul should not have all config details yet")
		}

		err = testRemoteConfig.GetConsul().AddKeyValue(common.StorageType, []byte("postgres"))
		if err != nil {
			return
		}
		creds, err = testRemoteConfig.GetStorageCreds(&storageType)
		if err == nil {
			t.Error("This call should fail. Consul should not have all config details yet")
		}

		// Setting everything at once. Key-value (static) postgresql creds
		err = SetStoragePostgres(testRemoteConfig.GetConsul().(*consul.Consulet),
			testRemoteConfig.GetVault(),
			"postgres",
			"localhost",
			fmt.Sprintf("%d", port),
			"postgres",
			pw,
			"kv",
			"")
		if err != nil {
			t.Error("Got an error trying to configure storage")
		}
		//println(testRemoteConfig)

		creds, err = testRemoteConfig.GetStorageCreds(&storageType)
		if err != nil {
			t.Error("Failed to get postgres storageCreds")
		}

		fmt.Println(creds)

		//// FIXME! From here on, expected that the vault instance is written against the Vagrant instance
		/// Need to change environment variables to use GetInstance()
		_ = os.Setenv("VAULT_ADDR", "http://192.168.56.78:8200")
		_ = os.Setenv("VAULT_TOKEN", "ocelotdev")

		consulHack := strings.Split(testRemoteConfig.GetConsul().(*consul.Consulet).Config.Address, ":")
		consulHackPort, _ := strconv.Atoi(consulHack[1])

		vagrantRemoteConfig, err := GetInstance(consulHack[0], consulHackPort, os.Getenv("VAULT_TOKEN"))
		if err != nil {
			t.Error("Failed trying to initialize vagrant remote config")
		}

		vaultDbConfig := "vault write database/config/ocelot " +
			"plugin_name=postgresql-database-plugin " +
			"allowed_roles='ocelot' " +
			"connection_url='postgresql://{{username}}:{{password}}@192.168.56.78:5432/?sslmode=disable' " +
			"username='postgres' " +
			"password='mysecretpassword'"

		vaultDbRole := `vault write database/roles/ocelot \
        db_name=ocelot \
        creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
            GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
        default_ttl="10m" \
        max_ttl="1h"`

		fmt.Printf("Vault config! %s\n", vaultDbConfig)
		fmt.Printf("Vault role! %s\n", vaultDbRole)

		// Set up test vault with the command line
		cmd := exec.Command("sh", "-c", fmt.Sprintf("vault secrets enable database || true"))
		//cmd := exec.Command("env")
		cmd.Env = append(cmd.Env, fmt.Sprintf("VAULT_ADDR=%s", "http://192.168.56.78:8200"))
		cmd.Env = append(cmd.Env, fmt.Sprintf("VAULT_TOKEN=%s", "ocelotdev"))

		stdoutStderr, err := cmd.CombinedOutput()
		if err != nil {
			t.Error("Got an error trying to set vault up via cli")
			fmt.Printf("I failed\n%s\n", stdoutStderr)
		} else {
			fmt.Printf("I passed\n%s\n", stdoutStderr)
		}

		///

		// Set up test vault with the command line
		cmd = exec.Command("sh", "-c", fmt.Sprintf("%s", vaultDbConfig))
		//cmd := exec.Command("env")
		cmd.Env = append(cmd.Env, fmt.Sprintf("VAULT_ADDR=%s", "http://192.168.56.78:8200"))
		cmd.Env = append(cmd.Env, fmt.Sprintf("VAULT_TOKEN=%s", "ocelotdev"))

		stdoutStderr, err = cmd.CombinedOutput()
		if err != nil {
			t.Error("Got an error trying to set vault up via cli")
			fmt.Printf("I failed\n%s\n", stdoutStderr)
		} else {
			fmt.Printf("I passed\n%s\n", stdoutStderr)
		}

		///

		// Set up test vault with the command line
		cmd = exec.Command("sh", "-c", fmt.Sprintf("%s", vaultDbRole))
		//cmd := exec.Command("env")
		cmd.Env = append(cmd.Env, fmt.Sprintf("VAULT_ADDR=%s", "http://192.168.56.78:8200"))
		cmd.Env = append(cmd.Env, fmt.Sprintf("VAULT_TOKEN=%s", "ocelotdev"))

		stdoutStderr, err = cmd.CombinedOutput()
		if err != nil {
			t.Error("Got an error trying to set vault up via cli")
			fmt.Printf("I failed\n%s\n", stdoutStderr)
		} else {
			fmt.Printf("I passed\n%s\n", stdoutStderr)
		}

		// FIXME: This is commented out until we can configure Vault to enable database backend + set up role against through testutil
		//// Re-configure consul + vault for dynamic postgres creds
		err = SetStoragePostgres(vagrantRemoteConfig.GetConsul().(*consul.Consulet),
			vagrantRemoteConfig.GetVault(),
			"postgres",
			"192.168.56.78",
			fmt.Sprintf("%d", 5432),
			"postgres",
			"mysecretpassword",
			"database",
			"ocelot")
		if err != nil {
			t.Error("Got an error trying to configure vault+postgres with dynamic creds")
		}
		println(vagrantRemoteConfig)

		storageType, err = vagrantRemoteConfig.GetStorageType()
		creds, err = vagrantRemoteConfig.GetStorageCreds(&storageType)
		if err != nil {
			t.Error("Failed to get postgres storageCreds")
		}

		fmt.Println(creds)

	case storage.FileSystem:
		t.Error("The expected storageType is postgres")
	default:
		t.Error("This should be an unreachable case - filesystem case")
	}

}

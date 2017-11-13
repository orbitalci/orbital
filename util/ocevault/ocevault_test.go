package ocevault

import (
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/vault"
	"github.com/shankj3/ocelot/util"
	"net"
	"os"
	"testing"
)

var userdata = []struct{
	username string
	testdata string
}{
	{"jessi", "7"},
	{"marianne", "17"},
	{"vardaofthevalier", "lotr"},
	{"test4", "17"},
}

func testSetupVaultAndAuthClient(t *testing.T) (oce *Ocevault, ln net.Listener) {
	core, _, token := vault.TestCoreUnsealed(t)
	ln, addr := http.TestServer(t, core)
	os.Setenv("VAULT_ADDR", addr)
	os.Setenv("VAULT_TOKEN", token)

	oce, err := NewEnvAuthClient()
	if err != nil {
		t.Fatalf("Could not init Auth client! Error: %s", err)
	}
	for _, ud := range userdata {
		datamap := make(map[string]interface{})
		datamap["test"] = ud.testdata
		_, err = oce.AddUserAuthData(ud.username, datamap)
		if err != nil {
			t.Fatal(err)
		}
	}
	return
}

func TestOcevault_CreateThrowawayToken(t *testing.T) {
	oce, ln := testSetupVaultAndAuthClient(t)
	defer ln.Close()
	// create Throwaway token
	secret, err := oce.CreateThrowawayToken()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("New Token:", secret)

	// make sure it can only be used once
	newCli, err := NewAuthedClient(secret)
	if err != nil {
		t.Fatalf("couldn't auth client with new token %s", err)
	}
	_, err = newCli.GetUserAuthData("marianne")
	if err != nil {
		t.Fatalf("Couldn't retrieve any data on first use. Error: %s",  err)
	}
	_, err = newCli.GetUserAuthData("jessi")
	if err == nil {
		t.Fatalf("Supposed to be single use token!")
	}
}
//
func TestOcevault_GetUserAuthData(t *testing.T) {
	oce, ln := testSetupVaultAndAuthClient(t)
	defer ln.Close()

	for _, ud := range userdata {
		sec, err := oce.GetUserAuthData(ud.username)
		if err != nil {
			t.Errorf("Unable to retrieve secret for user %s", ud.username)
		} else {
			if sec["test"] != ud.testdata {
				t.Error(util.GenericStrFormatErrors("UserAuthData", ud.testdata, sec["test"]))
			}
		}
	}
}

func TestOcevault_AddUserAuthData(t *testing.T) {
	var newdata = []struct {
		user string
		data interface{}
	}{
		{"user1", "17"},
		{"user7", "107"},
		{"user10", "23"},
		{"userweird", "weirdboy"},
	}
	oce, ln := testSetupVaultAndAuthClient(t)
	defer ln.Close()
	for _, nd := range newdata {
		data := make(map[string]interface{})
		data["test"] = nd.data
		_, err := oce.AddUserAuthData(nd.user, data)
		if err != nil {
			t.Errorf("could not add user data for %s, test data value %s, \n Error: %s", nd.user, nd.data, err)
		}
	}

	// now make sure you can access it
	for _, nd := range newdata {
		sec, err := oce.GetUserAuthData(nd.user)
		if err != nil {
			t.Errorf("Unable to get auth data for user %s \n Error: %s", nd.user, err)
		} else if sec["test"] != nd.data {
			t.Error(util.GenericStrFormatErrors("User Auth data Add", nd.data, sec["test"]))
		}
	}
}

//
//func TestOcevault_CreateOcevaultPolicy(t *testing.T) {
//	oce, err := NewEnvAuthClient()
//	if err != nil {
//		t.Fatal(err)
//	}
//	if err = oce.CreateOcevaultPolicy(); err != nil {
//		t.Fatal(err)
//	}
//}

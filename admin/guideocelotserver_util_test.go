package admin

import (
	"bitbucket.org/level11consulting/go-til/test"
	am "bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"testing"
)

func TestSetupRCCCredentials(t *testing.T) {
	// test all implementations of cred.RemoteConfigCred
	testRemoteConfig, vaultListener, consulServer := cred.TestSetupVaultAndConsul(t)
	defer cred.TeardownVaultAndConsul(vaultListener, consulServer)
	kubeconf := &am.K8SCreds{
		AcctName: "test",
		K8SContents: "laksjdflkjasdlkfjaeifnasd,mcnxo8r23$%(*asdf,zxddfh9\n\n\n\n\n\t\t\tdaslkdr73d8n!@#@!",
	}
	err := SetupRCCCredentials(testRemoteConfig, kubeconf)
	if err != nil {
		t.Fatal(err)
	}
	data, err := testRemoteConfig.GetVault().GetVaultData("secret/creds/k8s/test/k8s")
	if err != nil {
		t.Fatal(err)
	}
	v, ok := data["clientsecret"]
	if !ok {
		t.Fatal("unable to get clientsecret")
	}
	secret, ok := v.(string)
	if !ok {
		t.Fatal("unable to cast to string")
	}
	if secret != kubeconf.K8SContents {
		t.Error(test.StrFormatErrors("contents", kubeconf.K8SContents, secret))
	}
	vcsCreds := &am.VCSCreds{
		ClientId: "123",
		ClientSecret: "secsecsecret",
		TokenURL: "herebeaurl",
		AcctName: "herebeanaccount",
		Type: "bitbucket",
	}
	err = SetupRCCCredentials(testRemoteConfig, vcsCreds)
	if err != nil {
		t.Fatal(err)
	}
	data, err = testRemoteConfig.GetVault().GetVaultData("secret/creds/vcs/herebeanaccount/bitbucket")
	if err != nil {
		t.Fatal(err)
	}
	v, ok = data["clientsecret"]
	if !ok {
		t.Fatal("unable to get clientsecret")
	}
	if v != vcsCreds.ClientSecret {
		t.Error(test.StrFormatErrors("secret", vcsCreds.ClientSecret, v.(string)))
	}

}
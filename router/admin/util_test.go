package admin

import (
	"bitbucket.org/level11consulting/go-til/test"

	cred "bitbucket.org/level11consulting/ocelot/common/credentials"
	am "bitbucket.org/level11consulting/ocelot/models/pb"
	"bitbucket.org/level11consulting/ocelot/storage"

	"testing"
)

type dummyCredTable struct {
	storage.CredTable
}

func (d *dummyCredTable) InsertCred(credder am.OcyCredder, overWriteOk bool) error {
	return nil
}


func TestSetupRCCCredentials(t *testing.T) {
	// test all implementations of cred.RemoteConfigCred
	testRemoteConfig, vaultListener, consulServer := cred.TestSetupVaultAndConsul(t)
	defer cred.TeardownVaultAndConsul(vaultListener, consulServer)
	kubeconf := &am.K8SCreds{
		AcctName: "test",
		SubType:  am.SubCredType_KUBECONF,
		K8SContents: "laksjdflkjasdlkfjaeifnasd,mcnxo8r23$%(*asdf,zxddfh9\n\n\n\n\n\t\t\tdaslkdr73d8n!@#@!",
	}
	err := SetupRCCCredentials(testRemoteConfig, &dummyCredTable{}, kubeconf)
	if err != nil {
		t.Fatal(err)
	}
	data, err := testRemoteConfig.GetVault().GetVaultData("secret/creds/k8s/test/kubeconf")
	if err != nil {
		t.Fatal(err)
	}
	v, ok := data["data"]
	if !ok {
		t.Fatal("unable to get clientsecret")
	}
	q, ok := v.(map[string]interface{})
	if !ok {
		t.Fatal("wasnt a map??")
	}
	sec, ok := q["clientsecret"]
	if !ok {
		t.Fatal("wasn't a map??")
	}
	secret, ok := sec.(string)
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
		SubType: am.SubCredType_BITBUCKET,
	}
	err = SetupRCCCredentials(testRemoteConfig, &dummyCredTable{}, vcsCreds)
	if err != nil {
		t.Fatal(err)
	}
	data, err = testRemoteConfig.GetVault().GetVaultData("secret/creds/vcs/herebeanaccount/bitbucket")
	if err != nil {
		t.Fatal(err)
	}
	v, ok = data["data"]
	if !ok {
		t.Fatal("unable to get clientsecret")
	}
	q, ok = v.(map[string]interface{})
	if !ok {
		t.Fatal("wasnt a map??")
	}
	sec, ok = q["clientsecret"]
	if !ok {
		t.Fatal("wasn't a map??")
	}
	secret, ok = sec.(string)
	if !ok {
		t.Fatal("unable to cast to string")
	}
	if secret != vcsCreds.ClientSecret {
		t.Error(test.StrFormatErrors("secret", vcsCreds.ClientSecret, secret))
	}

}
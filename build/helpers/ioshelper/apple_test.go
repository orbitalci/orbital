package ioshelper

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/level11consulting/ocelot/common"
)

func TestAppleKeychain_GetSecretsFromZip(t *testing.T) {
	zipData, err := ioutil.ReadFile("./test-fixtures/test.zip")
	if err != nil {
		t.Fatal(err)
	}
	keychain := NewKeychain()
	err = keychain.GetSecretsFromZip(bytes.NewReader(zipData), "pw")
	if err != nil {
		t.Fatal(err)
	}
	aProfile, err := ioutil.ReadFile("./test-fixtures/developer/profiles/prof1.mobileprovision")
	if err != nil {
		t.Fatal(err)
	}
	encoded := common.BitzToBase64(aProfile)
	if mobile, ok := keychain.MobileProvisions["prof1.mobileprovision"]; !ok {
		t.Error("prof1.mobileprovision should exist in map")
	} else {
		if mobile != encoded {
			t.Error("returned mobile profile is not the same as the read one")
		}
	}
	identity, err := ioutil.ReadFile("./test-fixtures/developer/identities/id1.p12")
	if err != nil {
		t.Fatal(err)
	}
	encoded = common.BitzToBase64(identity)
	if profile, ok := keychain.PrivateKeys["id1.p12"]; !ok {
		t.Error("id1.p12 should exist in map")
	} else {
		if encoded != profile {
			t.Error("returned id1 not the same as read & encoded id1")
		}
	}
}

func TestUnpackAppleDevAccount(t *testing.T) {
	zipData, err := ioutil.ReadFile("./test-fixtures/test.zip")
	if err != nil {
		t.Fatal(err)
	}
	encodedTestData := GetAndEncodeDevFolder(t)
	expected := &AppleKeychain{
		PrivateKeys: map[string]string{"id1.p12": encodedTestData["id1.p12"]},
		MobileProvisions: map[string]string{
			"prof1.mobileprovision": encodedTestData["prof1.mobileprovision"],
			"prof2.mobileprovision": encodedTestData["prof2.mobileprovision"],
		},
		DevProfilePassword: "pw",
	}
	var expectedBits []byte
	expectedBits, err = json.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}
	liveBits, err := UnpackAppleDevAccount(zipData, "pw")
	if err != nil {
		t.Fatal("couldn't unzip dev account", err)
	}
	if !bytes.Equal(expectedBits, liveBits) {
		t.Log(string(expectedBits))
		t.Log(string(liveBits))
		t.Error("marshaled json not equal")
	}

}

package xcode

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/shankj3/ocelot/common"
	"github.com/shankj3/ocelot/models/pb"
)

func TestAppleDevProfile_GenerateIntegrationString(t *testing.T) {
	zipData, err := ioutil.ReadFile("./test-fixtures/test.zip")
	if err != nil {
		t.Fatal(err)
	}
	aProfile, err := ioutil.ReadFile("./test-fixtures/developer/profiles/0ee11910-8825-47e3-9ade-1ce6a7387fb6.mobileprovision")
	if err != nil {
		t.Fatal(err)
	}
	creder := &pb.AppleCreds{Identifier: "testytestytestytesty", AppleSecrets: zipData}
	creds := []pb.OcyCredder{creder}
	appledevprof := Create()
	appledevprof.GenerateIntegrationString(creds)

	unzippedBinaryFile, ok := appledevprof.zippedEncodedProfiles["testytestytestytesty"+idFilePathSeparator+"developer/profiles/0ee11910-8825-47e3-9ade-1ce6a7387fb6.mobileprovision"]
	if !ok {
		for key := range appledevprof.zippedEncodedProfiles {
			t.Log(key)
		}
		t.Error("testytestytestytesty"+ idFilePathSeparator + "developer/profiles/0ee11910-8825-47e3-9ade-1ce6a7387fb6.mobileprovision should exist in zippedEncodedProfiles")
	}
	binaryFileBitz, err := common.Base64ToBitz(unzippedBinaryFile)
	if err != nil {
		t.Error(err)
		return
	}
	if diff := bytes.Compare(binaryFileBitz, aProfile); diff != 0 {
		t.Log(string(binaryFileBitz))
		t.Log(string(aProfile))
		t.Log("but why though")
		t.Error("bytes not the same!!!  :'(")
		return
	}

}
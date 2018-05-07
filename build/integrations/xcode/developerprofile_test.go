package xcode

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/shankj3/ocelot/build/integrations"
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
	zipDString := string(zipData)
	creder := &pb.K8SCreds{Identifier: "testytestytestytesty", K8SContents: zipDString,}
	creds := []pb.OcyCredder{creder}
	adp := NewAppleDevProfile()
	adp.GenerateIntegrationString(creds)
	unzippedBinaryFile, ok := adp.zippedEncodedProfiles["testytestytestytesty"+idFilePathSeparator+"developer/profiles/0ee11910-8825-47e3-9ade-1ce6a7387fb6.mobileprovision"]
	if !ok {
		for key := range adp.zippedEncodedProfiles {
			t.Log(key)
		}
		t.Error("testytestytestytesty"+ idFilePathSeparator + "developer/profiles/0ee11910-8825-47e3-9ade-1ce6a7387fb6.mobileprovision should exist in zippedEncodedProfiles")
	}
	binaryFileBitz, err := integrations.Base64ToBitz(unzippedBinaryFile)
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
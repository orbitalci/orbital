package xcode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/go-test/deep"
	"github.com/level11consulting/ocelot/build/helpers/ioshelper"
	"github.com/level11consulting/ocelot/common"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/shankj3/go-til/test"
)

func TestAppleDevProfile_GenerateIntegrationString(t *testing.T) {
	aProfile, err := ioutil.ReadFile("./test-fixtures/test.mobileprovision")
	if err != nil {
		t.Fatal(err)
	}
	keychain := &ioshelper.AppleKeychain{MobileProvisions: map[string]string{"test.mobileprovision": common.BitzToBase64(aProfile)}, PrivateKeys: make(map[string]string)}
	marshaled, err := json.Marshal(keychain)
	if err != nil {
		t.Fatal(err)
	}
	creder := &pb.AppleCreds{Identifier: "testytestytestytesty", AppleSecrets: marshaled}
	creds := []pb.OcyCredder{creder}
	appledevprof := Create()
	appledevprof.GenerateIntegrationString(creds)
	envs := appledevprof.GetEnv()
	if len(envs) > 1 || envs[0] != "test_OCY_mobileprovision="+common.BitzToBase64(aProfile) {
		t.Error("only one env should have been rendered, and it should have been the base64 encoded mobile provisioning profile, test.mobileprovision")
	}
}

func TestAppleDevProfile_staticstuffs(t *testing.T) {
	prf := Create()
	if prf.String() != "apple dev profile integration" {
		t.Error("string rep should be apple dev profile integration")
	}
	if prf.SubType() != pb.SubCredType_DEVPROFILE {
		t.Error("subcredtype of apple dev profile should be DEVPROFILE")
	}
	wc := &pb.BuildConfig{BuildTool: "xcode"}
	if prf.IsRelevant(wc) {
		t.Error("xcode should be disabled")
	}
	wc = &pb.BuildConfig{BuildTool: "maven"}
	if prf.IsRelevant(wc) {
		t.Error("build tool is maven, this integration is not relevant")
	}
}

func TestAppleDevProfile_GetEnv(t *testing.T) {
	devProfile := Create()
	keyc := ioshelper.NewKeychain()
	zipMap := ioshelper.GetAndEncodeDevFolder(t)
	//test.zip is copied from common/ioshelper/test-fixtures, and has id1.p12, prof1.mobileprovision, prof2.mobileprovision
	zipBits, err := ioutil.ReadFile("./test-fixtures/test.zip")
	if err != nil {
		t.Fatal(err)
	}
	keyc.GetSecretsFromZip(bytes.NewReader(zipBits), "pw")
	devProfile.keys = []*ioshelper.AppleKeychain{keyc}
	generatedEnvs := devProfile.GetEnv()
	expectedEnvs := []string{
		fmt.Sprintf("id1_OCY_p12=%s", zipMap["id1.p12"]),
		fmt.Sprintf("prof1_OCY_mobileprovision=%s", zipMap["prof1.mobileprovision"]),
		fmt.Sprintf("prof2_OCY_mobileprovision=%s", zipMap["prof2.mobileprovision"]),
	}
	if diff := deep.Equal(expectedEnvs, generatedEnvs); diff != nil {
		t.Error(diff)
	}
}

func TestAppleDevProfile_MakeBashable(t *testing.T) {
	devProfile := &AppleDevProfile{joiner: "\n", pass: "TESTPASS"}
	keyc := ioshelper.NewKeychain()
	//test.zip is copied from common/ioshelper/test-fixtures, and has id1.p12, prof1.mobileprovision, prof2.mobileprovision
	zipBits, err := ioutil.ReadFile("./test-fixtures/test.zip")
	if err != nil {
		t.Fatal(err)
	}
	keyc.GetSecretsFromZip(bytes.NewReader(zipBits), "DEVTESTPASS")
	devProfile.keys = []*ioshelper.AppleKeychain{keyc}
	rendered := devProfile.MakeBashable("")
	expected, err := ioutil.ReadFile("./test-fixtures/expected.sh")
	if err != nil {
		t.Fatal(nil)
	}
	if rendered[0] != string(expected) {
		t.Error(test.StrFormatErrors("rendered bash", string(expected), rendered[0]))
	}
}

package dockerconfig

import (
	"github.com/go-test/deep"
	"github.com/shankj3/go-til/test"
	"github.com/shankj3/ocelot/models/pb"

	"bytes"
	"testing"
)

func Test_RCtoDockerConfig(t *testing.T) {
	repos := []pb.OcyCredder{&pb.RepoCreds{
		Username:   "mysuserisgr8",
		Password:   "apluspassword",
		RepoUrl:    "derp.docker.io",
		Identifier: "derpy",
		SubType:    pb.SubCredType_DOCKER,
	},
		&pb.RepoCreds{
			Username:   "whambam",
			Password:   "pw1237unsafe",
			RepoUrl:    "herp.docker.io",
			Identifier: "herpy",
			SubType:    pb.SubCredType_DOCKER,
		},
	}

	expected := []byte(`{"auths":{"derp.docker.io":{"auth":"bXlzdXNlcmlzZ3I4OmFwbHVzcGFzc3dvcmQ="},"herp.docker.io":{"auth":"d2hhbWJhbTpwdzEyMzd1bnNhZmU="}},"HttpHeaders":{"User-Agent":"Docker-Client/17.12.0-ce (linux)"}}`)
	jsonbit, err := RCtoDockerConfig(repos)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(expected, jsonbit) {
		t.Error("rendered docker config", string(expected), string(jsonbit))
	}
	//fmt.Println(string(jsonbit))
	//fmt.Println(base64.StdEncoding.EncodeToString(jsonbit))
}


func Test_RCtoDockerConfigFail(t *testing.T) {
	vcss := []pb.OcyCredder{&pb.VCSCreds{}}
	d := Create()
	_, err := d.GenerateIntegrationString(vcss)
	if err == nil {
		t.Error("should fail, did not pass repo creds")
	}
	if err.Error() != "unable to cast as repo creds" {
		t.Error("err msg", "unable to cast as repo creds", err.Error())
	}
}

func TestDockrInt_staticsstuffs(t *testing.T) {
	d := Create()
	di := d.(*DockrInt)
	if !d.IsRelevant(&pb.BuildConfig{}) {
		t.Error("docker config integration should always be relevant")
	}
	di.dConfig = ";alksdfjalk;sdjfklajsdfkl;ajsdfl;ajsdfl;aksdjfkl;ajsdfkl;asjdfl;jadslafjk"
	if diff := deep.Equal(d.GetEnv(), []string{"DCONF="+di.dConfig}); diff != nil {
		t.Error("Envs not equal, diff is: \n", diff)
	}
	expectedBashable := []string{"/bin/sh", "-c", "mkdir -p ~/.docker && echo \"${DCONF}\" | base64 -d > ~/.docker/config.json"}
	if diff := deep.Equal(d.MakeBashable("hummunu"), expectedBashable); diff != nil {
		t.Error("rendered bash strings not what they should be, diff is: ", diff)
	}
	if di.SubType() != pb.SubCredType_DOCKER {
		t.Error("subtype should be docker")
	}
	if di.String() != "docker login" {
		t.Error("string() should return 'docker login'")
	}

}

func TestDockrInt_GenerateIntegrationString(t *testing.T) {
	repos := []pb.OcyCredder{&pb.RepoCreds{
		Username:   "mysuserisgr8",
		Password:   "apluspassword",
		RepoUrl:    "derp.docker.io",
		Identifier: "derpy",
		SubType:    pb.SubCredType_DOCKER,
	},
		&pb.RepoCreds{
			Username:   "whambam",
			Password:   "pw1237unsafe",
			RepoUrl:    "herp.docker.io",
			Identifier: "herpy",
			SubType:    pb.SubCredType_DOCKER,
		},
	}
	di := Create()
	configjson, err := di.GenerateIntegrationString(repos)
	if err != nil {
		t.Error(err)
	}
	expectedEncoded := "eyJhdXRocyI6eyJkZXJwLmRvY2tlci5pbyI6eyJhdXRoIjoiYlhsemRYTmxjbWx6WjNJNE9tRndiSFZ6Y0dGemMzZHZjbVE9In0sImhlcnAuZG9ja2VyLmlvIjp7ImF1dGgiOiJkMmhoYldKaGJUcHdkekV5TXpkMWJuTmhabVU9In19LCJIdHRwSGVhZGVycyI6eyJVc2VyLUFnZW50IjoiRG9ja2VyLUNsaWVudC8xNy4xMi4wLWNlIChsaW51eCkifX0="
	if configjson != expectedEncoded {
		t.Error(test.StrFormatErrors("encoded docker config.json", expectedEncoded, configjson))
	}
}

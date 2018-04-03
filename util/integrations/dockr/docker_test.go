package dockr

import (
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bytes"
	//"encoding/base64"
	//"fmt"
	"testing"
)

func Test_RCtoDockerConfig(t *testing.T) {
	repo := &models.RepoCreds{
		Username: "mysuserisgr8",
		Password: "apluspassword",
		RepoUrl: map[string]string{
			"derpy": "derp.docker.io",
			"herpy": "herp.docker.io",
		},
	}
	expected := []byte(`{"auths":{"derp.docker.io":{"auth":"bXlzdXNlcmlzZ3I4OmFwbHVzcGFzc3dvcmQ="},"herp.docker.io":{"auth":"bXlzdXNlcmlzZ3I4OmFwbHVzcGFzc3dvcmQ="}},"HttpHeaders":{"User-Agent":"Docker-Client/17.12.0-ce (linux)"}}`)
	jsonbit, err := RCtoDockerConfig(repo)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(expected, jsonbit) {
		t.Error("rendered docker config", string(expected), string(jsonbit))
	}
	//fmt.Println(string(jsonbit))
	//fmt.Println(base64.StdEncoding.EncodeToString(jsonbit))
}
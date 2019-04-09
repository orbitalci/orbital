/*
 dockerconfig is an implementation of the StringIntegrator interface

	Its methods will grab docker repo credentials and transform them into docker config.json files that will be authentication for
	docker repos. If docker repo creds are added to ocelot for the VCS account that is building, then the build will have access to images in
	that docker repo.
*/
package dockerconfig

import (
	"github.com/level11consulting/ocelot/models/pb"

	"encoding/json"
	"errors"
	"fmt"

	"github.com/level11consulting/ocelot/build/helpers/serde"
)

type auth map[string]string

type dockerConfigJson struct {
	Auths       map[string]auth   `json:"auths,omitempty"`
	HttpHeaders map[string]string `json:"HttpHeaders,omitempty"`
}

type DockrInt struct {
	dConfig string
}

func (d *DockrInt) String() string {
	return "docker login"
}

func (d *DockrInt) SubType() pb.SubCredType {
	return pb.SubCredType_DOCKER
}

func (d *DockrInt) GenerateIntegrationString(credz []pb.OcyCredder) (string, error) {
	bitz, err := RCtoDockerConfig(credz)
	if err != nil {
		return "", err
	}
	configEncoded := serde.BitzToBase64(bitz)
	d.dConfig = configEncoded
	return configEncoded, err
}

func (d *DockrInt) MakeBashable(encoded string) []string {
	return []string{"/bin/sh", "-c", "mkdir -p ~/.docker && echo \"${DCONF}\" | base64 -d > ~/.docker/config.json"}
}

func (d *DockrInt) IsRelevant(wc *pb.BuildConfig) bool {
	return true
}

func (d *DockrInt) GetEnv() []string {
	return []string{"DCONF=" + d.dConfig}
}

func RCtoDockerConfig(creds []pb.OcyCredder) ([]byte, error) {
	authz := make(map[string]auth)
	for _, credi := range creds {
		credx, ok := credi.(*pb.RepoCreds)
		if !ok {
			return nil, errors.New("unable to cast as repo creds")
		}
		authstring := fmt.Sprintf("%s:%s", credx.Username, credx.Password)
		b64authstring := serde.StrToBase64(authstring)
		authz[credx.RepoUrl] = map[string]string{"auth": b64authstring}
	}
	config := &dockerConfigJson{
		Auths:       authz,
		HttpHeaders: map[string]string{"User-Agent": "Docker-Client/17.12.0-ce (linux)"},
	}
	bitz, err := json.Marshal(config)
	return bitz, err
}

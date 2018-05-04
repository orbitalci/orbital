package dockerconfig

import (
	"github.com/shankj3/ocelot/build/integrations"
	"github.com/shankj3/ocelot/models/pb"

	"encoding/json"
	"errors"
	"fmt"
)

type auth map[string]string

type dockerConfigJson struct {
	Auths       map[string]auth   `json:"auths,omitempty"`
	HttpHeaders map[string]string `json:"HttpHeaders,omitempty"`
}

type DockrInt struct{}

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
	configEncoded := integrations.BitzToBase64(bitz)
	return configEncoded, err
}

func (d *DockrInt) MakeBashable(encoded string) []string {
	return []string{"/bin/sh", "-c", "/.ocelot/render_docker.sh " + "'" + encoded + "'"}
}

func (d *DockrInt) IsRelevant(wc *pb.BuildConfig) bool {
	return true
}

func (d *DockrInt) GetEnv() []string {
	return []string{}
}

func RCtoDockerConfig(creds []pb.OcyCredder) ([]byte, error) {
	authz := make(map[string]auth)
	for _, credi := range creds {
		credx, ok := credi.(*pb.RepoCreds)
		if !ok {
			return nil, errors.New("unable to cast as repo creds")
		}
		authstring := fmt.Sprintf("%s:%s", credx.Username, credx.Password)
		b64authstring := integrations.StrToBase64(authstring)
		authz[credx.RepoUrl] = map[string]string{"auth": b64authstring}
	}
	config := &dockerConfigJson{
		Auths:       authz,
		HttpHeaders: map[string]string{"User-Agent": "Docker-Client/17.12.0-ce (linux)"},
	}
	bitz, err := json.Marshal(config)
	return bitz, err
}

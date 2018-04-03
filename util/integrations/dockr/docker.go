package dockr

import (
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/integrations"
	"encoding/json"
	"errors"
	"fmt"
)


type auth map[string]string

type dockerConfigJson struct {
	Auths map[string]auth `json:"auths,omitempty"`
	HttpHeaders map[string]string `json:"HttpHeaders,omitempty"`

}

// GetDockerConfig will find docker creds associated with accountName in the CVRemoteConfig, and will
// generate a config.json authentication file for docker. The contents will be returned base64 encoded for
// easy passing as a command line argument.
func GetDockerConfig(rc cred.CVRemoteConfig, accountName string) (string, error) {
	repod := models.NewRepoCreds()
	credz, err := rc.GetCredAt(fmt.Sprintf(cred.Docker, accountName), false, repod)
	if err != nil {
		return "", err
	}
	dockerCred, ok := credz[cred.BuildCredKey("docker", accountName)]
	if !ok {
		return "", integrations.NCErr("no creds found")
	}
	casted, ok := dockerCred.(*models.RepoCreds)
	if !ok {
		return "", errors.New(fmt.Sprintf("unable to cast to RepoCreds, which just shouldn't happen. Object: %v", dockerCred))
	}
	bitz, err := RCtoDockerConfig(casted)
	if err != nil {
		return "", err
	}
	configEncoded := integrations.BitzToBase64(bitz)
	return configEncoded, err
}


func RCtoDockerConfig(creds *models.RepoCreds) ([]byte, error) {
	authstring := fmt.Sprintf("%s:%s", creds.Username, creds.Password)
	b64authstring := integrations.StrToBase64(authstring)
	authz := make(map[string]auth)
	for _, url := range creds.RepoUrl {
		authz[url] = map[string]string{"auth":b64authstring}
	}
	config := &dockerConfigJson{
		Auths: authz,
		HttpHeaders: map[string]string{"User-Agent": "Docker-Client/17.12.0-ce (linux)"},
	}
	bitz, err := json.Marshal(config)
	return bitz, err
}


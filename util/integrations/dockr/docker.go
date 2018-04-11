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
	//GetCredsBySubTypeAndAcct(stype pb.SubCredType, accountName string, hideSecret bool)
	credz, err := rc.GetCredsBySubTypeAndAcct(models.SubCredType_DOCKER, accountName, false)
	if err != nil {
		return "", err
	}
	bitz, err := RCtoDockerConfig(credz)
	if err != nil {
		return "", err
	}
	configEncoded := integrations.BitzToBase64(bitz)
	return configEncoded, err
}


func RCtoDockerConfig(creds []models.OcyCredder) ([]byte, error) {
	authz := make(map[string]auth)
	for _, credi := range creds {
		credx, ok := credi.(*models.RepoCreds)
		if !ok {
			return nil, errors.New("unable to cast as repo creds")
		}
		authstring := fmt.Sprintf("%s:%s", credx.Username, credx.Password)
		b64authstring := integrations.StrToBase64(authstring)
		authz[credx.RepoUrl] = map[string]string{"auth":b64authstring}
	}
	config := &dockerConfigJson{
		Auths: authz,
		HttpHeaders: map[string]string{"User-Agent": "Docker-Client/17.12.0-ce (linux)"},
	}
	bitz, err := json.Marshal(config)
	return bitz, err
}


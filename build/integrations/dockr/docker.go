package dockr

import (
	"bitbucket.org/level11consulting/ocelot/build/integrations"
	cred "bitbucket.org/level11consulting/ocelot/common/credentials"
	"bitbucket.org/level11consulting/ocelot/models/pb"
	"bitbucket.org/level11consulting/ocelot/storage"

	"encoding/json"
	"errors"
	"fmt"
)


type auth map[string]string

type dockerConfigJson struct {
	Auths map[string]auth `json:"auths,omitempty"`
	HttpHeaders map[string]string `json:"HttpHeaders,omitempty"`

}

type DockrInt struct {}

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
//
//func (d *DockrInt) GetThemCreds(rc cred.CVRemoteConfig, store storage.CredTable, accountName string) ([]pb.OcyCredder, error) {
//	credz, err := rc.GetCredsBySubTypeAndAcct(store, pb.SubCredType_DOCKER, accountName, false)
//	if err != nil {
//		return nil, err
//	}
//	return credz, nil
//}


// GetDockerConfig will find docker creds associated with accountName in the CVRemoteConfig, and will
// generate a config.json authentication file for docker. The contents will be returned base64 encoded for
// easy passing as a command line argument.
func GetDockerConfig(rc cred.CVRemoteConfig, store storage.CredTable, accountName string) (string, error) {
	//GetCredsBySubTypeAndAcct(stype pb.SubCredType, accountName string, hideSecret bool)
	credz, err := rc.GetCredsBySubTypeAndAcct(store, pb.SubCredType_DOCKER, accountName, false)
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


func RCtoDockerConfig(creds []pb.OcyCredder) ([]byte, error) {
	authz := make(map[string]auth)
	for _, credi := range creds {
		credx, ok := credi.(*pb.RepoCreds)
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


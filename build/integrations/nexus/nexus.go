package nexus

import (
	"bitbucket.org/level11consulting/ocelot/build/integrations"
	cred "bitbucket.org/level11consulting/ocelot/common/credentials"
	"bitbucket.org/level11consulting/ocelot/models/pb"
	"bitbucket.org/level11consulting/ocelot/storage"

	"bytes"
	"errors"
	"text/template"
)

var settingsXml = `<?xml version="1.0" encoding="UTF-8"?>
<settings xmlns="http://maven.apache.org/SETTINGS/1.1.0"
  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
  xsi:schemaLocation="http://maven.apache.org/SETTINGS/1.1.0 http://maven.apache.org/xsd/settings-1.1.0.xsd">
  <servers>
	{{range .Repo}}
    <server>
      <id>{{.Identifier}}</id>
      <username>{{.Username}}</username>
      <password>{{.Password}}</password>
    </server>
	{{ end }}
  </servers>
  <profiles>
    <profile>
      <id>level11consulting</id>
      <activation>
        <activeByDefault>true</activeByDefault>
      </activation>
      <repositories>
        <repository>
          <id>ocelotNexus</id>
          <name>Ocelot Rendered</name>
          <url>${env.NEXUS_PUBLIC_M2}</url>
        </repository>
      </repositories>
    </profile>
  </profiles>
</settings>`

type NexusInt struct {}

func (n *NexusInt) String() string {
	return "nexus m2 settings.xml render"
}

func (n *NexusInt) SubType() pb.SubCredType {
	return pb.SubCredType_NEXUS
}

func Create() integrations.StringIntegrator {
	return &NexusInt{}
}


func (n *NexusInt) GenerateIntegrationString(credz []pb.OcyCredder) (string, error) {
	var repoCreds []*pb.RepoCreds
	for _, credi := range credz {
		credx, ok := credi.(*pb.RepoCreds)
		if !ok {
			return "", errors.New("could not cast as repo creds")
		}
		repoCreds = append(repoCreds, credx)
	}
	wrap := &pb.RepoCredWrapper{Repo:repoCreds}
	return executeTempl(wrap)
}

// GetSettingsXml will render and return a maven settings.xml with credentials correlating to the accountName provided
// todo: include project name for further filtering
func GetSettingsXml(rc cred.CVRemoteConfig, store storage.CredTable, accountName string) (string, error) {
	credz, err := rc.GetCredsBySubTypeAndAcct(store, pb.SubCredType_NEXUS, accountName, false)
	if err != nil {
		return "", err
	}
	var repoCreds []*pb.RepoCreds
	for _, credi := range credz {
		credx, ok := credi.(*pb.RepoCreds)
		if !ok {
			return "", errors.New("could not cast as repo creds")
		}
		repoCreds = append(repoCreds, credx)
	}
	wrap := &pb.RepoCredWrapper{Repo:repoCreds}
	return executeTempl(wrap)
}

func executeTempl(wrap *pb.RepoCredWrapper) (string, error) {
	templ, err := template.New("settingsxml").Parse(settingsXml)
	if err != nil {
		return "", err
	}
	var settings bytes.Buffer
	err = templ.Execute(&settings, wrap)
	if err != nil {
		return "", errors.New("unable to render settings.xml template for nexus credentials. error: " + err.Error())
	}
	return settings.String(), nil

}
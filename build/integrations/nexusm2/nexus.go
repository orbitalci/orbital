package nexusm2

import (
	"bitbucket.org/level11consulting/ocelot/build/integrations"
	"bitbucket.org/level11consulting/ocelot/models/pb"

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

type NexusInt struct {
	settingsXml string
}

func (n *NexusInt) String() string {
	return "nexus m2 settings.xml render"
}

func (n *NexusInt) SubType() pb.SubCredType {
	return pb.SubCredType_NEXUS
}

func Create() integrations.StringIntegrator {
	return &NexusInt{}
}

func (n *NexusInt) GetEnv() []string {
	return []string{"M2XML=" + n.settingsXml}
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
	rendered, err := executeTempl(wrap)
	if err == nil {
		n.settingsXml = rendered
	}
	return rendered, err
}

func (n *NexusInt) MakeBashable(xml string) []string {
	return []string{"/bin/sh", "-c", "mkdir -p ~/.m2 && echo \"${M2XML}\" > ~/.m2/settings.xml"}
}

func (n *NexusInt) IsRelevant(wc *pb.BuildConfig) bool {
	if wc.BuildTool == "maven" {
		return true
	}
	return false
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
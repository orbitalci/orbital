package nexus

import (
	"bitbucket.org/level11consulting/ocelot/old/admin/models"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/storage"

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


// GetSettingsXml will render and return a maven settings.xml with credentials correlating to the accountName provided
// todo: include project name for further filtering
func GetSettingsXml(rc cred.CVRemoteConfig, store storage.CredTable, accountName string) (string, error) {
	credz, err := rc.GetCredsBySubTypeAndAcct(store, models.SubCredType_NEXUS, accountName, false)
	if err != nil {
		return "", err
	}
	var repoCreds []*models.RepoCreds
	for _, credi := range credz {
		credx, ok := credi.(*models.RepoCreds)
		if !ok {
			return "", errors.New("could not cast as repo creds")
		}
		repoCreds = append(repoCreds, credx)
	}
	wrap := &models.RepoCredWrapper{Repo:repoCreds}
	return executeTempl(wrap)
}

func executeTempl(wrap *models.RepoCredWrapper) (string, error) {
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
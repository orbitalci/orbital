package nexus

import (
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bytes"
	"errors"
	"fmt"
	"text/template"
)

const settingsXml = `<?xml version="1.0" encoding="UTF-8"?>
<settings xmlns="http://maven.apache.org/SETTINGS/1.1.0"
  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
  xsi:schemaLocation="http://maven.apache.org/SETTINGS/1.1.0 http://maven.apache.org/xsd/settings-1.1.0.xsd">
  <servers>
  {{ $name, _ := range .RepoUrl }}
    <server>
      <id>{{$name}}</id>
      <username>{{.Username}}</username>
      <password>{{.Password}}</password>
    </server>
  {{ end }}
  </servers>
</settings>`

var templ = template.Must(template.ParseGlob(settingsXml))


func GetSettingsXml(rc cred.CVRemoteConfig, accountName string) (string, error) {
	repo := models.NewRepoCreds()
	credz, err := rc.GetCredAt(fmt.Sprintf(cred.Nexus, accountName), false, repo)
	if err != nil {
		return "", err
	}
	nexusCred, ok := credz[cred.BuildCredKey("nexus", accountName)]
	if !ok {
		return "", errors.New("could not find nexus credentials in remote config")
	}
	casted, ok := nexusCred.(*models.RepoCreds)
	if !ok {
		return "", errors.New("unable to cast to REpoCreds, which just shouldn't happen")
	}
	var settings bytes.Buffer
	err = templ.Execute(&settings, casted)
	if err != nil {
		return "", errors.New("unable to render settings.xml template for nexus credentials. error: " + err.Error())
	}
	return settings.String(), nil
}
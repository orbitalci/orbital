package nexusm2

import (
	"testing"

	"github.com/shankj3/go-til/test"
	"github.com/level11consulting/ocelot/models/pb"
)

var expected = `<?xml version="1.0" encoding="UTF-8"?>
<settings xmlns="http://maven.apache.org/SETTINGS/1.1.0"
  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
  xsi:schemaLocation="http://maven.apache.org/SETTINGS/1.1.0 http://maven.apache.org/xsd/settings-1.1.0.xsd">
  <servers>
    
    <server>
      <id>myFirstRepo</id>
      <username>testuser1</username>
      <password>testpw</password>
    </server>
    
    <server>
      <id>mySecondRepo</id>
      <username>testuser2</username>
      <password>testpw2</password>
    </server>
    
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

func Test_executeTempl(t *testing.T) {
	creds := []*pb.RepoCreds{
		{Username: "testuser1", Password: "testpw", RepoUrl: "testRepo.com", Identifier: "myFirstRepo"},
		{Username: "testuser2", Password: "testpw2", RepoUrl: "11testRepo.com", Identifier: "mySecondRepo"},
	}
	wrap := &pb.RepoCredWrapper{Repo: creds}
	template, err := executeTempl(wrap)
	if err != nil {
		t.Error(err)
	}
	if expected != template {
		t.Error("should be the same?")
	}
}

func TestNexusInt_GenerateIntegrationString(t *testing.T) {
	creds := []pb.OcyCredder{
		&pb.RepoCreds{Username: "testuser1", Password: "testpw", RepoUrl: "testRepo.com", Identifier: "myFirstRepo"},
		&pb.RepoCreds{Username: "testuser2", Password: "testpw2", RepoUrl: "11testRepo.com", Identifier: "mySecondRepo"},
	}
	integ := Create()
	rendered, err := integ.GenerateIntegrationString(creds)
	if err != nil {
		t.Error(err)
		return
	}
	if rendered != expected {
		t.Error(test.StrFormatErrors("rendered settings.xml", expected, rendered))
	}
}

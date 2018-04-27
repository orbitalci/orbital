package build

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/shankj3/ocelot/models/pb"
)

func TestOcelotValidator_ValidateConfig(t *testing.T) {
	goodconfig := &pb.BuildConfig{
		Image:     "busybox:latest",
		BuildTool: "maven",
		Packages:  []string{},
		Branches:  []string{"ALL"},
		Env:       []string{},
		Stages:    []*pb.Stage{{Name: "one"}, {Name: "build"}},
	}
	valid8r := GetOcelotValidator()
	err := valid8r.ValidateConfig(goodconfig, nil)
	if err != nil {
		t.Error(err)
	}
	badConfig := &pb.BuildConfig{
		Image:     "busybox:latest",
		BuildTool: "maven",
		Packages:  []string{},
		Branches:  []string{"ALL"},
		Env:       []string{},
		Stages:    []*pb.Stage{{Name: "one"}},
	}
	err = valid8r.ValidateConfig(badConfig, nil)
	if err == nil {
		t.Error("should not be nil, as there is no build stage")
	}
	// todo: flag this as integration test
	// privateRepo should be something that you have creds to
	privateRepo, ok := os.LookupEnv("PRIVATE_REGISTRY")
	if !ok {
		t.Log("using default privateRepo of docker.metaverse.l11.com")
		privateRepo = "docker.metaverse.l11.com"
	}
	dckrconfig, err := ioutil.ReadFile(os.ExpandEnv("$HOME/.docker/config.json"))
	if err != nil {
		t.Log("skipping nexus pull test because couldn't get docker config")
		return
	}
	if !strings.Contains(string(dckrconfig), privateRepo) {
		t.Log("skipping nexus pull test because " + privateRepo + " is not in docker config")
		return
	}
	goodPrivateConfig := &pb.BuildConfig{
		Image:     privateRepo + "/busybox:test_do_not_delete",
		BuildTool: "maven",
		Packages:  []string{},
		Branches:  []string{"ALL"},
		Env:       []string{},
		Stages:    []*pb.Stage{{Name: "one"}, {Name: "build"}},
	}
	err = valid8r.ValidateConfig(goodPrivateConfig, nil)
	if err != nil {
		t.Log("this may have failed becasue " + privateRepo + "/busybox:test_do_not_delete is not in metaverse anymore")
		t.Error(err)
	}

}

package dockrhelper

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)
func TestRobustImagePull(t *testing.T) {
	out, err := RobustImagePull("busybox:latest")
	out.Close()
	if err != nil {
		t.Error("should have pulled successfully, error: ", err.Error())
	}
	if !pulledByApi {
		t.Error("should have pulled via the docker api and returned as there are no creds required")
	}
	out, err = RobustImagePull("garbage89123b8")
	defer out.Close()
	if err == nil {
		t.Error("should not successfully have pulled a dummy image")
	}
}

func TestRobustImagePull_integration(t *testing.T) {
	// todo: flag this as integration test
	// privateRepo should be something that you have creds to
	privateRepo, ok := os.LookupEnv("PRIVATE_REGISTRY")
	if !ok {
		t.Log("using default privateRepo of docker.metaverse.l11.com")
		privateRepo = "docker.metaverse.l11.com"
	}
	dckrconfig, err := ioutil.ReadFile(os.ExpandEnv("$HOME/.docker/config.json"))
	if err != nil {
		t.Skip("skipping nexus pull test because couldn't get docker config")
	}
	if !strings.Contains(string(dckrconfig), privateRepo) {
		t.Skip("skipping nexus pull test because " + privateRepo + " is not in docker config")
	}
	privateImage := privateRepo + "/busybox:test_do_not_delete"
	out, err := RobustImagePull(privateImage)
	defer out.Close()
	if err != nil {
		t.Error("may have failed because " + privateImage + "has been deleted when it shouldn't have been. error: ", err.Error())
		return
	}
}
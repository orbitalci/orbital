package helm

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/shankj3/ocelot/models/pb"
)


func TestCreate(t *testing.T) {
	h := Create("localhost", "9090")
	hel := h.(*helmInteg)
	if hel.ip != "localhost" || hel.port != "9090" {
		t.Error("helmInteg not properly instantiated")
	}
}

func TestHelmInteg_GenerateDownloadBashables(t *testing.T) {
	h := Create("localhost", "9090")
	expectedShell := []string{"/bin/sh", "-c", "cd /tmp && wget http://localhost:9090/helm.tar.gz && tar -xzvf helm.tar.gz && mv linux-amd64/helm /bin"}
	shells := h.GenerateDownloadBashables()
	if diff := deep.Equal(expectedShell, shells); diff != nil {
		t.Error(diff)
	}
}

func TestHelmInteg_IsRelevant(t *testing.T) {
	wc := &pb.BuildConfig{
		Stages: []*pb.Stage{
			{Script: []string{"mvn clean install"}},
			{Script: []string{"mkdir sohely", "cd sohely", "helm update"}},
		},
	}
	h := Create("localhost", "9090")
	if !h.IsRelevant(wc) {
		t.Error("should be relevant, as 'helm' exists in the script")
	}
}

func TestHelmInteg_String(t *testing.T) {
	h := Create("localhost", "9090")
	if h.String() != "helm binary downloader" {
		t.Error("its dumb that i have to test this for code coverage.")
	}
}

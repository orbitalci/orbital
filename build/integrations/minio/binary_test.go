package minio

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/shankj3/ocelot/models/pb"
)

func TestCreate(t *testing.T) {
	h := Create("localhost", "9090")
	hel := h.(*minioInteg)
	if hel.ip != "localhost" || hel.port != "9090" {
		t.Error("minioInteg not properly instantiated")
	}
}

func TestMinioInteg_GenerateDownloadBashables(t *testing.T) {
	h := Create("localhost", "9090")
	expectedShell := []string{"/bin/sh", "-c", "cd /bin && wget http://localhost:9090/mc && chmod +x mc"}
	shells := h.GenerateDownloadBashables()
	if diff := deep.Equal(expectedShell, shells); diff != nil {
		t.Error(diff)
	}
}

func TestMinioInteg_IsRelevant(t *testing.T) {
	wc := &pb.BuildConfig{
		Stages: []*pb.Stage{
			{Script: []string{"mvn clean install"}},
			{Script: []string{"mkdir sohely", "cd sohely", "mc upload"}},
		},
	}
	h := Create("localhost", "9090")
	if !h.IsRelevant(wc) {
		t.Error("should be relevant, as 'helm' exists in the script")
	}
}

func TestMinioInteg_String(t *testing.T) {
	h := Create("localhost", "9090")
	if h.String() != "minio cli (mc) binary downloader" {
		t.Error("its dumb that i have to test this for code coverage.")
	}
}

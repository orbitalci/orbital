package kubectl

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/level11consulting/ocelot/models/pb"
)

func TestCreate(t *testing.T) {
	h := Create("localhost", "9090")
	hel := h.(*kubectlInteg)
	if hel.ip != "localhost" || hel.port != "9090" {
		t.Error("KubectlInteg not properly instantiated")
	}
}

func TestKubectlInteg_GenerateDownloadBashables(t *testing.T) {
	h := Create("localhost", "9090")
	expectedShell := []string{"/bin/sh", "-c", "cd /bin && wget http://localhost:9090/kubectl && chmod +x kubectl"}
	shells := h.GenerateDownloadBashables()
	if diff := deep.Equal(expectedShell, shells); diff != nil {
		t.Error(diff)
	}
}

func TestKubectlInteg_IsRelevant(t *testing.T) {
	wc := &pb.BuildConfig{
		Stages: []*pb.Stage{
			{Script: []string{"mvn clean install"}},
			{Script: []string{"mkdir sohely", "cd sohely", "kubectl apply -f"}},
		},
	}
	h := Create("localhost", "9090")
	if !h.IsRelevant(wc) {
		t.Error("should be relevant, as 'kubectl' exists in the script")
	}
}

func TestKubectlInteg_String(t *testing.T) {
	h := Create("localhost", "9090")
	if h.String() != "kubectl binary downloader" {
		t.Error("its dumb that i have to test this for code coverage.")
	}
}

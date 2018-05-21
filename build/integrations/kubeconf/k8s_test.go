package kubeconf

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/shankj3/go-til/test"
	"github.com/shankj3/ocelot/models/pb"
)

func TestK8sInt_GenerateIntegrationString(t *testing.T) {
	inte := &K8sInt{}
	conf := []pb.OcyCredder{&pb.K8SCreds{
		K8SContents: "wasssuppppppp",
		Identifier:  "derpy",
		SubType:     pb.SubCredType_KUBECONF,
	},
	}
	kubeconf, err := inte.GenerateIntegrationString(conf)
	if err != nil {
		t.Error(err)
		return
	}
	expected := "d2Fzc3N1cHBwcHBwcA=="
	if expected != kubeconf {
		t.Error(test.StrFormatErrors("rendered kubeconf", expected, kubeconf))
	}

	badcreds := []pb.OcyCredder{&pb.RepoCreds{}}
	_, err = inte.GenerateIntegrationString(badcreds)
	if err == nil {
		t.Error("should return error as GenerateIntegrationString was passed RepoCreds")
	}
	if err.Error() != "could not cast to k8s cred" {
		t.Error(test.StrFormatErrors("err msg", "could not cast to k8s cred", err.Error()))
	}
}


func TestK8sInt_staticstuffs(t *testing.T) {
	shouldbeK8s := Create()
	inte, ok := shouldbeK8s.(*K8sInt)
	if !ok {
		t.Error("Create() didn't return a k8s int obj")
	}
	if inte.String() != "kubeconfig render" {
		t.Error(test.StrFormatErrors("string() value", "kubeconfig render", inte.String()))
	}
	if inte.SubType() != pb.SubCredType_KUBECONF {
		t.Errorf("subtype should be KUBECONF, got %s", inte.SubType())
	}
	if !inte.IsRelevant(&pb.BuildConfig{}) {
		t.Error("kubeconfig render is currently always relevant")
	}
	inte.k8sContents = "a;lsdfjkal;skdfjakl;sdfj"
	if diff := deep.Equal(inte.GetEnv(), []string{"KCONF=" + inte.k8sContents}); diff != nil {
		t.Error("getEnv not right, diff is: \n", diff)
	}
	expectedbash := []string{"/bin/sh", "-c", "mkdir -p ~/.kube && echo \"${KCONF}\" | base64 -d > ~/.kube/conf"}
	if diff := deep.Equal(expectedbash, inte.MakeBashable("a;lsdfjk")); diff != nil {
		t.Error("bash strings not rendered correctly, diff is: \n", diff)
	}

}
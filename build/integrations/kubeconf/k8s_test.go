package kubeconf

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/shankj3/go-til/test"
	"github.com/shankj3/ocelot/models/pb"
)

// Test cases: 0-2 integrations
func TestK8sInt_GenerateIntegrationString(t *testing.T) {
	inte := &K8sInt{}
	// Zero integrations
	conf := []pb.OcyCredder{}
	kubeconf, err := inte.GenerateIntegrationString(conf)
	if err != nil {
		t.Error(err)
		return
	}
	expected := "e30="
	if expected != kubeconf {
		t.Error(test.StrFormatErrors("rendered kubeconf json", expected, kubeconf))
	}

	// One integration
	conf = append(conf, &pb.K8SCreds{
		K8SContents: "wasssuppppppp",
		Identifier:  "derpy",
		SubType:     pb.SubCredType_KUBECONF,
	})

	kubeconf, err = inte.GenerateIntegrationString(conf)
	if err != nil {
		t.Error(err)
		return
	}
	expected = "eyJkZXJweSI6Indhc3NzdXBwcHBwcHAifQ=="
	if expected != kubeconf {
		t.Error(test.StrFormatErrors("rendered kubeconf json", expected, kubeconf))
	}

	// Two integrations
	conf = append(conf, &pb.K8SCreds{
		K8SContents: "such digital, very amaze",
		Identifier:  "doge",
		SubType:     pb.SubCredType_KUBECONF,
	})
	kubeconf, err = inte.GenerateIntegrationString(conf)
	if err != nil {
		t.Error(err)
		return
	}
	expected = "eyJkZXJweSI6Indhc3NzdXBwcHBwcHAiLCJkb2dlIjoic3VjaCBkaWdpdGFsLCB2ZXJ5IGFtYXplIn0="
	if expected != kubeconf {
		t.Error(test.StrFormatErrors("rendered kubeconf json", expected, kubeconf))
	}
}

// FIXME! Runtime panic, pointer dereference
//func TestK8sInt_InvalidCredsGenerateIntegrationString(t *testing.T) {
//  inte := &K8sInt{}
//	badcreds := []pb.OcyCredder{&pb.RepoCreds{}}
//	_, err := inte.GenerateIntegrationString(badcreds)
//	if err == nil {
//		t.Error("should return error as GenerateIntegrationString was passed RepoCreds")
//	}
//	if err.Error() != "could not cast to k8s cred" {
//		t.Error(test.StrFormatErrors("err msg", "could not cast to k8s cred", err.Error()))
//	}
//}

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
	inte.k8sContents = "eyJkZXJweSI6Indhc3NzdXBwcHBwcHAifQ=="
	if diff := deep.Equal(inte.GetEnv(), []string{"derpy=wasssuppppppp", "K8S_INDEX=derpy "}); diff != nil {
		t.Error("getEnv not right, diff is: \n", diff)
	}
	expectedbash := []string{"/bin/bash", "-c", "mkdir -p ~/.kube && for kubeconf in ${K8S_INDEX}; do echo \"${!kubeconf}\" > ~/.kube/${kubeconf}; done"}
	if diff := deep.Equal(expectedbash, inte.MakeBashable("a;lsdfjk")); diff != nil {
		t.Error("bash strings not rendered correctly, diff is: \n", diff)
	}

}

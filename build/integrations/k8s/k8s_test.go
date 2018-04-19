package k8s

import (
	"testing"

	"bitbucket.org/level11consulting/go-til/test"
	"bitbucket.org/level11consulting/ocelot/models/pb"
)

func TestK8sInt_GenerateIntegrationString(t *testing.T) {
	inte := &K8sInt{}
	conf := []pb.OcyCredder{&pb.K8SCreds{
		K8SContents: "wasssuppppppp",
		Identifier: "derpy",
		SubType:    pb.SubCredType_KUBECONF,
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
}
package helmrepo

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/shankj3/go-til/test"
	"github.com/shankj3/ocelot/models/pb"
)

var conf = []pb.OcyCredder{
	&pb.GenericCreds{
		Identifier:   "myhelmrepo",
		ClientSecret: "http://helm.me",
		SubType:      pb.SubCredType_HELM_REPO,
		AcctName:     "schmear",
	}, &pb.GenericCreds{
		Identifier:   "thehelmrepo",
		ClientSecret: "http://helm.co",
		SubType:      pb.SubCredType_HELM_REPO,
		AcctName:     "schmear",
	},
}

func TestHelmRepoInt_GenerateIntegrationString_MakeBashable(t *testing.T) {
	integ := &HelmRepoInt{}
	helmRepoInt, err := integ.GenerateIntegrationString(conf)
	if err != nil {
		t.Error(err)
	}
	expected := `helm init --client-only; helm repo add myhelmrepo http://helm.me; helm repo add thehelmrepo http://helm.co; `
	if expected != helmRepoInt {
		t.Error(test.StrFormatErrors("rendered helm commands,", expected, helmRepoInt))
	}

	helmRepoInt, err = integ.GenerateIntegrationString([]pb.OcyCredder{conf[1]})
	if err != nil {
		t.Error(err)
	}
	expected = `helm init --client-only; helm repo add thehelmrepo http://helm.co; `
	if expected != helmRepoInt {
		t.Error(test.StrFormatErrors("rendered helm commands,", expected, helmRepoInt))
	}
	cmds := integ.MakeBashable(helmRepoInt)
	if diff := deep.Equal(cmds, []string{"/bin/bash", "-c", helmRepoInt}); diff != nil {
		t.Error(diff)
	}
}

func TestHelmRepoInt_Statics(t *testing.T) {
	helmy := Create()
	if helmy.String() != "helm repo configuration" {
		t.Error(test.StrFormatErrors("string", "helm repo configuration", helmy.String()))
	}
	if helmy.SubType() != pb.SubCredType_HELM_REPO {
		t.Error(test.StrFormatErrors("subtype", pb.SubCredType_HELM_REPO.String(), helmy.SubType().String()))
	}
	if diff := deep.Equal(helmy.GetEnv(), []string{}); diff != nil {
		t.Error(diff)
	}
}

func TestHelmRepoInt_IsRelevant(t *testing.T) {
	helmy := &HelmRepoInt{}
	wc := &pb.BuildConfig{
		Stages: []*pb.Stage{
			{Script: []string{"helm"}},
		},
	}
	if !helmy.IsRelevant(wc) {
		t.Error("has helm in script, should return true")
	}
}

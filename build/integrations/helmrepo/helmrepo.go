package helmrepo

import (
	"fmt"

	"github.com/shankj3/ocelot/build/integrations"
	"github.com/shankj3/ocelot/common"
	"github.com/shankj3/ocelot/models/pb"
)

func Create() integrations.StringIntegrator {
	return &HelmRepoInt{}
}

type HelmRepoInt struct {}

func (k *HelmRepoInt) String() string {
	return "helm repo configuration"
}

func (k *HelmRepoInt) SubType() pb.SubCredType {
	return pb.SubCredType_HELM_REPO
}

// Create a config file for every k8s entry, named ("~/.kube/%s", k.identifier)
func (k *HelmRepoInt) MakeBashable(encoded string) []string {
	return []string{"/bin/bash", "-c", encoded}
}


func (k *HelmRepoInt) IsRelevant(wc *pb.BuildConfig) bool {
	return common.BuildScriptsContainString(wc, "helm")
}

func (k *HelmRepoInt) GenerateIntegrationString(creds []pb.OcyCredder) (string, error) {
	var repoAddString string
	repoAddString = `helm init --client-only; `
	for _, helmrepo := range creds {
		repoAddString += fmt.Sprintf("helm repo add %s %s; ", helmrepo.GetIdentifier(), helmrepo.GetClientSecret())
	}
	return repoAddString, nil
}

// Deserialize k.k8sContents, and build multiple environment varialbes
func (k *HelmRepoInt) GetEnv() []string {
	return []string{}
}

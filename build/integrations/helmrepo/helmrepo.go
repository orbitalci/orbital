/*
  helmrepo is an implementation of the StringIntegrator interface

	If there is "helm" in any of the commands in the ocelot.yml file and there are helm repo credentials, then `helm init` will be run
	and all the helm repos will be added to the container's local helm index
*/
package helmrepo

import (
	"fmt"

	"github.com/level11consulting/ocelot/build/integrations"
	"github.com/level11consulting/ocelot/common"
	"github.com/level11consulting/ocelot/models/pb"
)

func Create() integrations.StringIntegrator {
	return &HelmRepoInt{}
}

type HelmRepoInt struct{}

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

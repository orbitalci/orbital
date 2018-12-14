/*
  kubeconf is an implementation of the StringIntegrator interface

	Its methods will use all the the kubeconfig credentials to generate kubeconfig files in the ~/.kube directory
*/

package kubeconf

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/level11consulting/orbitalci/build/helpers/serde"
	"github.com/level11consulting/orbitalci/build/integrations"
	"github.com/level11consulting/orbitalci/models/pb"
)

func Create() integrations.StringIntegrator {
	return &K8sInt{}
}

type K8sInt struct {
	k8sContents string
}

func (k *K8sInt) String() string {
	return "kubeconfig render"
}

func (k *K8sInt) SubType() pb.SubCredType {
	return pb.SubCredType_KUBECONF
}

// Create a config file for every k8s entry, named ("~/.kube/%s", k.identifier)
func (k *K8sInt) MakeBashable(encoded string) []string {
	//return []string{"/bin/bash", "-c", "mkdir -p ~/.kube && for i in ${K8S_INDEX}; do echo \"${!i}\" > ~/.kube/${i}; done && echo Printing file ~/.kube/${i} && cat ~/.kube/${i}"}
	return []string{"/bin/bash", "-c", "mkdir -p ~/.kube && for kubeconf in ${K8S_INDEX}; do echo \"${!kubeconf}\" > ~/.kube/${kubeconf}; done"}
}

// TODO: This integration is only relevant if we call kubectl or helm in any step
func (k *K8sInt) IsRelevant(wc *pb.BuildConfig) bool {
	return true
}

func (k *K8sInt) GenerateIntegrationString(creds []pb.OcyCredder) (string, error) {
	// Stuff creds into a map, and convert into json so we can pass it with some context for environment variables
	multiCreds := make(map[string]string)
	for _, cluster := range creds {
		multiCreds[cluster.GetIdentifier()] = cluster.GetClientSecret()
	}
	multiCredsJson, _ := json.Marshal(multiCreds)

	configEncoded := serde.StrToBase64(string(multiCredsJson))
	k.k8sContents = configEncoded
	return configEncoded, nil
}

// Deserialize k.k8sContents, and build multiple environment varialbes
func (k *K8sInt) GetEnv() []string {
	conf_json, _ := serde.Base64ToBitz(k.k8sContents)
	configs := make(map[string]string)
	json.Unmarshal([]byte(string(conf_json)), &configs)

	env_vars := make([]string, len(configs)+1)
	index := 0

	var cluster_list bytes.Buffer
	for cluster, conf := range configs {
		var buffer bytes.Buffer
		// FIXME: For backwards compatibility. Remove after allowing existing builds to migrate to new functionality
		if cluster == "THERECANONLYBEONE" {
			buffer.WriteString(fmt.Sprintf("%s=%s", "config", conf))
			cluster_list.WriteString(fmt.Sprintf("%s ", "config"))
		} else {
			buffer.WriteString(fmt.Sprintf("%s=%s", cluster, conf))
			cluster_list.WriteString(fmt.Sprintf("%s ", cluster))
		}
		env_vars[index] = buffer.String()
		index++
	}

	// We will use K8S_INDEX to reference the kubeconfig environment vars in a bash loop
	env_vars[index] = fmt.Sprintf("%s=%s", "K8S_INDEX", cluster_list.String())
	return env_vars
}

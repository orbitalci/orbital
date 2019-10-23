/*
  kubectl is an implementation of the BinaryIntegrator interface

	kubectl's methods will download  kubectl to the path of the docker container user if there is the string 'kubectl' in any of the
	stages' commands in the ocelot.yml file
*/
package kubectl

import (
	"fmt"

	"github.com/level11consulting/orbitalci/build/helpers/buildscript/search"
	"github.com/level11consulting/orbitalci/build/integrations"
	"github.com/level11consulting/orbitalci/models/pb"
)

func Create(loopbackip string, port string) integrations.BinaryIntegrator {
	return &kubectlInteg{ip: loopbackip, port: port}
}

type kubectlInteg struct {
	ip   string
	port string
}

func (k *kubectlInteg) GenerateDownloadBashables() []string {
	downloadLink := fmt.Sprintf("http://%s:%s/kubectl", k.ip, k.port)
	return []string{"/bin/sh", "-c", "cd /bin && wget " + downloadLink + " && chmod +x kubectl"}
}

func (k *kubectlInteg) IsRelevant(wc *pb.BuildConfig) bool {
	return search.BuildScriptsContainString(wc, "kubectl")
}

func (k *kubectlInteg) String() string {
	return "kubectl binary downloader"
}

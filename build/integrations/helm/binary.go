/*
  helm is an implementation of the BinaryIntegrator interface

	If there is a command the has the string "helm" in it, then the binary helm will be downloaded and installed in the build user's path.
*/
package helm

import (
	"fmt"

	"github.com/level11consulting/ocelot/build/integrations"
	"github.com/level11consulting/ocelot/common"
	"github.com/level11consulting/ocelot/models/pb"
)

func Create(loopbackip string, port string) integrations.BinaryIntegrator {
	return &helmInteg{ip: loopbackip, port: port}
}

type helmInteg struct {
	ip   string
	port string
}

func (k *helmInteg) GenerateDownloadBashables() []string {
	downloadLink := fmt.Sprintf("http://%s:%s/helm.tar.gz", k.ip, k.port)
	return []string{"/bin/sh", "-c", "cd /tmp && wget " + downloadLink + " && tar -xzvf helm.tar.gz && mv linux-amd64/helm /bin"}
}

func (k *helmInteg) IsRelevant(wc *pb.BuildConfig) bool {
	return common.BuildScriptsContainString(wc, "helm")
}

func (k *helmInteg) String() string {
	return "helm binary downloader"
}

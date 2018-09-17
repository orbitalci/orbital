package kubectl

import (
	"fmt"

	"github.com/shankj3/ocelot/build/integrations"
	"github.com/shankj3/ocelot/common"
	"github.com/shankj3/ocelot/models/pb"
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
	return common.BuildScriptsContainString(wc, "kubectl")
}

func (k *kubectlInteg) String() string {
	return "kubectl binary downloader"
}

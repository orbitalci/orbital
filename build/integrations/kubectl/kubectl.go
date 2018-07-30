package kubectl

import (
	"fmt"
	"strings"

	"github.com/shankj3/ocelot/build/integrations"
	"github.com/shankj3/ocelot/models/pb"
)

func Create(loopbackip string, port string) integrations.BinaryIntegrator {
	return &kubectlInteg{ip: loopbackip, port: port}
}

type kubectlInteg struct {
	ip string
	port string
}

func (k *kubectlInteg) GenerateDownloadBashables() []string {
	downloadLink := fmt.Sprintf("http://%s:%s/kubectl", k.ip, k.port)
	return []string{"/bin/sh", "-c", "cd /bin && wget " + downloadLink + " && chmod +x kubectl"}
}

func (k *kubectlInteg) IsRelevant(wc *pb.BuildConfig) bool {
	for _, stage := range wc.Stages {
		for _, script := range stage.Script {
			if strings.Contains(script, "kubectl") {
				return true
			}
		}
	}
	return false
}

func (k *kubectlInteg) String() string {
	return "kubectl binary downloader"
}
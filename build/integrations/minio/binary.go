/*
this package is for implementing the BinaryIntegrator interface to download the minio client for builds
*/

package minio

import (
	"fmt"

	"github.com/shankj3/ocelot/build/integrations"
	"github.com/shankj3/ocelot/common"
	"github.com/shankj3/ocelot/models/pb"
)

func Create(loopbackip string, port string) integrations.BinaryIntegrator {
	return &minioInteg{ip: loopbackip, port: port}
}

type minioInteg struct {
	ip string
	port string
}

func (k *minioInteg) GenerateDownloadBashables() []string {
	downloadLink := fmt.Sprintf("http://%s:%s/mc", k.ip, k.port)
	return []string{"/bin/sh", "-c", "cd /bin && wget " + downloadLink + " && chmod +x mc"}
}

func (k *minioInteg) IsRelevant(wc *pb.BuildConfig) bool {
	return common.BuildScriptsContainString(wc, "mc")
}

func (k *minioInteg) String() string {
	return "minio cli (mc) binary downloader"
}
package kubeconf

import (
	"errors"

	"github.com/shankj3/ocelot/build/integrations"
	"github.com/shankj3/ocelot/models/pb"
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

func (k *K8sInt) MakeBashable(encoded string) []string {
	return []string{"/bin/sh", "-c", "mkdir -p ~/.kube && echo \"${KCONF}\" | base64 -d > ~/.kube/conf"}
}

func (k *K8sInt) IsRelevant(wc *pb.BuildConfig) bool {
	return true
}

func (k *K8sInt) GenerateIntegrationString(creds []pb.OcyCredder) (string, error) {
	kubeCred, ok := creds[0].(*pb.K8SCreds)
	if !ok {
		return "", errors.New("could not cast to k8s cred")
	}
	configEncoded := integrations.StrToBase64(kubeCred.K8SContents)
	k.k8sContents = configEncoded
	return configEncoded, nil
}

func (k *K8sInt) GetEnv() []string {
	return []string{"KCONF=" + k.k8sContents}
}

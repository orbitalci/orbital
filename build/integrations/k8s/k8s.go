package k8s

import (
	"errors"

	cred "bitbucket.org/level11consulting/ocelot/common/credentials"
	"bitbucket.org/level11consulting/ocelot/build/integrations"
	"bitbucket.org/level11consulting/ocelot/models/pb"
	"bitbucket.org/level11consulting/ocelot/storage"
)

func Create() integrations.StringIntegrator {
	return &K8sInt{}
}

type K8sInt struct {}

func (k *K8sInt) String() string {
	return "kubeconfig render"
}

func (k *K8sInt) SubType() pb.SubCredType {
	return pb.SubCredType_KUBECONF
}

//func (k *K8sInt) GetThemCreds(rc cred.CVRemoteConfig, store storage.CredTable, accountName string) ([]pb.OcyCredder, error) {
//	credz, err := rc.GetCredsBySubTypeAndAcct(store, pb.SubCredType_KUBECONF, accountName, false)
//	if err != nil {
//		return nil, err
//	}
//	return credz, nil
//}

func (k *K8sInt) GenerateIntegrationString(creds []pb.OcyCredder) (string, error) {
	kubeCred, ok := creds[0].(*pb.K8SCreds)
	if !ok {
		return "", errors.New("could not cast to k8s cred")
	}
	configEncoded := integrations.StrToBase64(kubeCred.K8SContents)
	return configEncoded, nil
}

// GetKubeConfig will return a base64 encoded string of the kubeconfig
// kubeconfig is only supported for ONE KUBECONFIG PER ACCOUNT for now.
func GetKubeConfig(rc cred.CVRemoteConfig, store storage.CredTable, accountName string) (string, error) {

	credz, err := rc.GetCredsBySubTypeAndAcct(store, pb.SubCredType_KUBECONF, accountName, false)
	if err != nil {
		return "", err
	}
	
	kubeCred, ok := credz[0].(*pb.K8SCreds)
	if !ok {
		return "", errors.New("could not cast to k8s cred")
	}
	configEncoded := integrations.StrToBase64(kubeCred.K8SContents)
	return configEncoded, nil
}

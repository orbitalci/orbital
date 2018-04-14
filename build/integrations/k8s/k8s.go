package k8s

import (
	"errors"

	cred "bitbucket.org/level11consulting/ocelot/common/credentials"
	"bitbucket.org/level11consulting/ocelot/build/integrations"
	"bitbucket.org/level11consulting/ocelot/models/pb"
	"bitbucket.org/level11consulting/ocelot/storage"
)

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

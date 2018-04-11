package k8s

import (
	"errors"

	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/integrations"
)

// GetKubeConfig will return a base64 encoded string of the kubeconfig
// kubeconfig is only supported for ONE KUBECONFIG PER ACCOUNT for now.
func GetKubeConfig(rc cred.CVRemoteConfig, accountName string) (string, error) {

	credz, err := rc.GetCredsBySubTypeAndAcct(models.SubCredType_KUBECONF, accountName, false)
	if err != nil {
		return "", err
	}
	
	kubeCred, ok := credz[0].(*models.K8SCreds)
	if !ok {
		return "", errors.New("could not cast to k8s cred")
	}
	configEncoded := integrations.StrToBase64(kubeCred.K8SContents)
	return configEncoded, nil
}

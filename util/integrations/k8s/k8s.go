package k8s

import (
	"errors"
	"fmt"

	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/integrations"
)

// GetKubeConfig will return a base64 encoded string of the kubeconfig
func GetKubeConfig(rc cred.CVRemoteConfig, accountName string) (string, error) {
	k8s := models.NewK8sCreds()
	credz, err := rc.GetCredAt(fmt.Sprintf(cred.Kubernetes, accountName), false, k8s)
	if err != nil {
		return "", err
	}
	k8sCreds, ok := credz[cred.BuildCredKey("k8s", accountName)]
	if !ok {
		return "", integrations.NCErr("no creds found")
	}
	k8sCred, ok := k8sCreds.(*models.K8SCreds)
	if !ok {
		return "", errors.New("could not cast to k8s cred")
	}
	configEncoded := integrations.StrToBase64(k8sCred.K8SContents)
	return configEncoded, nil
}

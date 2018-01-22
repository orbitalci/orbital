package commandhelper

import (
	"bitbucket.org/level11consulting/ocelot/admin/models"
)

func NewTestClientConfig(logLines []string) *ClientConfig {
	return &ClientConfig{
		Client: models.NewFakeGuideOcelotClient(logLines),
	}
}
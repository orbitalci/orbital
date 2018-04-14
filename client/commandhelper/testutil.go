package commandhelper

import (
	"bitbucket.org/level11consulting/ocelot/old/admin/models"
)

func NewTestClientConfig(logLines []string) *ClientConfig {
	return &ClientConfig{
		Client: models.NewFakeGuideOcelotClient(logLines),
	}
}
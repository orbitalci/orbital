package commandhelper

import "bitbucket.org/level11consulting/ocelot/admin/models"

func NewTestClientConfig() *ClientConfig {
	return &ClientConfig{
		Client: models.NewFakeGuideOcelotClient(),
	}
}
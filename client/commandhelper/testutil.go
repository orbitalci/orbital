package commandhelper

import (
	"github.com/shankj3/ocelot/common/testutil"
)

func NewTestClientConfig(logLines []string) *ClientConfig {
	return &ClientConfig{
		Client: testutil.NewFakeGuideOcelotClient(logLines),
		Theme:  Default(false),
	}
}

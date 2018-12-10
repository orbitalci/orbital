package commandhelper

import (
	"github.com/level11consulting/ocelot/common/testutil"
)

func NewTestClientConfig(logLines []string) *ClientConfig {
	return &ClientConfig{
		Client: testutil.NewFakeGuideOcelotClient(logLines),
		Theme:  Default(false),
	}
}

package dockerconfig

import (
	"github.com/level11consulting/ocelot/build/integrations"
)

func Create() integrations.StringIntegrator {
	return &DockrInt{}
}

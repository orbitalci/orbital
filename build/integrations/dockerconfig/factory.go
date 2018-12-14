package dockerconfig

import (
	"github.com/level11consulting/orbitalci/build/integrations"
)

func Create() integrations.StringIntegrator {
	return &DockrInt{}
}

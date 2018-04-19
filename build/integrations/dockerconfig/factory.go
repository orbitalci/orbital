package dockerconfig

import (
	"bitbucket.org/level11consulting/ocelot/build/integrations"
)


func Create() integrations.StringIntegrator {
	return &DockrInt{}
}
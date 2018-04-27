package dockerconfig

import (
    "github.com/shankj3/ocelot/build/integrations"
)

func Create() integrations.StringIntegrator {
    return &DockrInt{}
}

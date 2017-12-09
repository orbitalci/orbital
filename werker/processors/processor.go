package processors

import (
	"github.com/shankj3/ocelot/protos"
)

type Processor interface {
	RunPushBundle(bund *protos.PushBuildBundle, infoChan chan []byte)
	RunPRBundle(bund *protos.PRBuildBundle, infoChan chan []byte)
}


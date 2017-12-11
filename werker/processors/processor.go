package processors

import (
    "bitbucket.org/level11consulting/ocelot/protos"
)

type Processor interface {
    RunPushBundle(bund *protos.PushBuildBundle, infoChan chan []byte)
    RunPRBundle(bund *protos.PRBuildBundle, infoChan chan []byte)
}

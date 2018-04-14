package launcher

import (
	"bitbucket.org/level11consulting/ocelot/build/basher"
	"bitbucket.org/level11consulting/ocelot/build/valet"
	"bitbucket.org/level11consulting/ocelot/common/credentials"
	"bitbucket.org/level11consulting/ocelot/models"
	"bitbucket.org/level11consulting/ocelot/storage"
)

// create a struct that is werker msg handler creates when it receives a new nsq message

type Launcher struct {
	*models.WerkerFacts
	Type        	models.WerkType
	RemoteConf     credentials.CVRemoteConfig
	infochan       chan []byte
	StreamChan   	chan *models.Transport
	BuildCtxChan 	chan *models.BuildContext
	Basher         *basher.Basher
	Store          storage.OcelotStorage
	BuildValet     *valet.Valet

}
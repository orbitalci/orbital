package secret

import (

	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/server/config"

	"github.com/level11consulting/ocelot/storage"

)

func SetupRCCCredentials(remoteConf config.CVRemoteConfig, store storage.CredTable, config pb.OcyCredder) error {
	//right now, we will always overwrite
	err := remoteConf.AddCreds(store, config, true)
	return err
}
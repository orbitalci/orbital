package legacy

import (
	"github.com/level11consulting/orbitalci/models/pb"
	"github.com/level11consulting/orbitalci/server/config"
	"github.com/level11consulting/orbitalci/storage"
)

func SetupRCCCredentials(remoteConf config.CVRemoteConfig, store storage.CredTable, config pb.OcyCredder) error {
	//right now, we will always overwrite
	err := remoteConf.AddCreds(store, config, true)
	return err
}

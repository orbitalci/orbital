package admin

import (

	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/server/config"

	"github.com/level11consulting/ocelot/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

)

func SetupRCCCredentials(remoteConf config.CVRemoteConfig, store storage.CredTable, config pb.OcyCredder) error {
	//right now, we will always overwrite
	err := remoteConf.AddCreds(store, config, true)
	return err
}

// handleStorageError  will attempt to decipher if err is not found. if so, iwll set the appropriate grpc status code and return new grpc status error
func HandleStorageError(err error) error {
	if _, ok := err.(*storage.ErrNotFound); ok {
		return status.Error(codes.NotFound, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}

package remote

import (
	"errors"

	"bitbucket.org/level11consulting/ocelot/common/remote/bitbucket"
	"bitbucket.org/level11consulting/ocelot/models"
	"bitbucket.org/level11consulting/ocelot/models/pb"
)

func GetHandler(creds *pb.VCSCreds) (handler models.VCSHandler, token string, err error) {
	switch creds.SubType {
	case pb.SubCredType_BITBUCKET:
		handler, token, err = bitbucket.GetBitbucketClient(creds)
	case pb.SubCredType_GITHUB:
		handler, token, err =  nil, "", errors.New("github not yet implemented")
	default:
		handler, token, err = nil, "", errors.New("subtype " + creds.SubType.String() + " not implemented")
	}
	return
}

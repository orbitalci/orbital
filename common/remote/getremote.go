package remote

import (
	"errors"
	"strings"

	"github.com/shankj3/ocelot/common/remote/bitbucket"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
)

func GetHandler(creds *pb.VCSCreds) (handler models.VCSHandler, token string, err error) {
	switch creds.SubType {
	case pb.SubCredType_BITBUCKET:
		handler, token, err = bitbucket.GetBitbucketClient(creds)
	case pb.SubCredType_GITHUB:
		handler, token, err = nil, "", errors.New("github not yet implemented")
	default:
		handler, token, err = nil, "", errors.New("subtype "+creds.SubType.String()+" not implemented")
	}
	return
}

// GetRemoteTranslator is for getting a models.Translator for converting webhook post bodies to pr/push events
func GetRemoteTranslator(sct pb.SubCredType) (models.Translator, error) {
	switch sct {
	case pb.SubCredType_BITBUCKET:
		return bitbucket.GetTranslator(), nil
	case pb.SubCredType_GITHUB:
		return nil, errors.New("github not currently suppported")
	default:
		return nil, errors.New("currently only bitbucket is supported for translation, recieved: " + sct.String())
	}
}

// GetVCSSubTypeFromUrl will parse the url given and figure out what vcs it originates from
func GetVCSSubTypeFromUrl(url string) (pb.SubCredType, error) {
	switch {
	case strings.Contains(url, "bitbucket"):
		return pb.SubCredType_BITBUCKET,  nil
	case strings.Contains(url, "github"):
		return pb.SubCredType_GITHUB, nil
	default:
		return 0, errors.New("unsupported vcs type")
	}
}
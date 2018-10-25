package remote

import (
	"context"
	"errors"
	"strings"

	"github.com/shankj3/ocelot/common/remote/bitbucket"
	"github.com/shankj3/ocelot/common/remote/github"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
	"golang.org/x/oauth2"
)

func GetHandler(creds *pb.VCSCreds) (handler models.VCSHandler, token string, err error) {
	switch creds.SubType {
	case pb.SubCredType_BITBUCKET:
		handler, token, err = bitbucket.GetBitbucketClient(creds)
	case pb.SubCredType_GITHUB:
		handler, token, err = github.GetGithubClient(creds)
	default:
		handler, token, err = nil, "", errors.New("subtype "+creds.SubType.String()+" not implemented")
	}
	return
}

// GetHandlerWithToken will create an oauth2 client with the token, then return a handler instantiated with that oauth2 client.
//   This client is NOT configured to autorenew, as it doesn't have the client secret required to do so. It assumes a *STATIC* token source!!
func GetHandlerWithToken(ctx context.Context, accessToken string, subType pb.SubCredType) (handler models.VCSHandler, err error) {
	token := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	authCli := oauth2.NewClient(ctx, token)
	switch subType {
	case pb.SubCredType_BITBUCKET:
		return bitbucket.GetBitbucketFromHttpClient(authCli), nil
	case pb.SubCredType_GITHUB:
		return github.GetGithubFromHttpClient(authCli), nil
	default:
		return nil, errors.New("unknown vcs type, cannot create handler with token given")
	}
}

// GetRemoteTranslator is for getting a models.Translator for converting webhook post bodies to pr/push events
func GetRemoteTranslator(sct pb.SubCredType) (models.Translator, error) {
	switch sct {
	case pb.SubCredType_BITBUCKET:
		return bitbucket.GetTranslator(), nil
	case pb.SubCredType_GITHUB:
		return github.GetTranslator(), nil
	default:
		return nil, errors.New("currently only bitbucket|github are supported for translation, recieved: " + sct.String())
	}
}

// GetVCSSubTypeFromUrl will parse the url given and figure out what vcs it originates from
func GetVCSSubTypeFromUrl(url string) (pb.SubCredType, error) {
	switch {
	case strings.Contains(url, "bitbucket"):
		return pb.SubCredType_BITBUCKET, nil
	case strings.Contains(url, "github"):
		return pb.SubCredType_GITHUB, nil
	default:
		return 0, errors.New("unsupported vcs type")
	}
}

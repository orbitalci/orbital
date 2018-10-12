package github

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"

	ocenet "github.com/shankj3/go-til/net"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
)

//Returns VCS handler for pulling source code and auth token if exists (auth token is needed for code download)
func GetGithubClient(creds *pb.VCSCreds) (models.VCSHandler, string, error) {
	client := &ocenet.OAuthClient{}
	token, err := client.Setup(creds)
	if err != nil {
		return nil, "", errors.New("unable to retrieve token for " + creds.AcctName + ".  Error: " + err.Error())
	}
	gh := GetGithubHandler(creds, client)
	return gh, token, nil
}

func GetGithubHandler(cred *pb.VCSCreds, cli ocenet.HttpClient) *Github {
	return &Github{
		Client:        cli,
		Marshaler:     jsonpb.Marshaler{},
		Unmarshaler:   jsonpb.Unmarshaler{AllowUnknownFields: true},
		credConfig:    cred,
		isInitialized: true,
	}
}

type Github struct {
	CallbackURL   string
	RepoBaseURL   string
	Client 		  ocenet.HttpClient
	Marshaler     jsonpb.Marshaler
	Unmarshaler   jsonpb.Unmarshaler
	credConfig    *pb.VCSCreds
	isInitialized bool

	models.VCSHandler
}

func (gh *Github) GetClient() ocenet.HttpClient {
	return gh.Client
}


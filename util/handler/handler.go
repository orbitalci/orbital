package handler

import (
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	ocenet "bitbucket.org/level11consulting/go-til/net"
)

type VCSHandler interface {
	//Walk will iterate over all repositories for specified vcs account, and create webhooks at specified webhook url
	//if one does not yet exist
	Walk() error

	//GetFile retrieves file based on file path, full repo name, and commit hash
	GetFile(filePath string, fullRepoName string, commitHash string) (bytez []byte, err error)

	//CreateWebhook will create a webhook using the webhook creation endpoint associated with codebase
	CreateWebhook(webhookURL string) error

	//GetCallbackURL retrieves the current callback URL
	GetCallbackURL() string

	//SetCallbackURL sets the callback URL for webhooks
	SetCallbackURL(callbackURL string)

	//SetBaseURL set the base URL for this handler
	SetBaseURL(baseURL string)

	//GetBaseURL returns the base URL for this handler
	GetBaseURL() string

	//FindWebhooks iterates over existing webhooks and returns true (matches our callback urls) if one already exists
	FindWebhooks(getWebhookURL string) bool

	//Get Repository details by account name + repo name
	GetRepoDetail(acctRepo string) (pb.PaginatedRepository_RepositoryValues, error)

	//Get repository's details by account name, repo, and hash
	GetHashDetail(acctRepo, hash string) (pb.ChangeSetV1, error)
}

//Returns VCS handler for pulling source code and auth token if exists (auth token is needed for code download)
func GetBitbucketClient(cfg *models.VCSCreds) (VCSHandler, string, error) {
	bbClient := &ocenet.OAuthClient{}
	token, err := bbClient.Setup(cfg)
	if err != nil {
		return nil, "", err
	}
	bb := GetBitbucketHandler(cfg, bbClient)
	return bb, token, nil
}
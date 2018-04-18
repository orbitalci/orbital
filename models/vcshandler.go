package models

import (
	// ugh stuck 4 now
	pbb "bitbucket.org/level11consulting/ocelot/models/bitbucket/pb"
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
	GetRepoDetail(acctRepo string) (pbb.PaginatedRepository_RepositoryValues, error)

	//GetAllCommits returns a paginated list of commits corresponding with branch
	GetAllCommits(acctRepo string, branch string) (*pbb.Commits, error)
}
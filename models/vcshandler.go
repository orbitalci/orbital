package models

import (
	"fmt"
	"io"

	ocenet "github.com/shankj3/go-til/net"
	"github.com/shankj3/ocelot/models/pb"
)

//go:generate mockgen -source vcshandler.go -destination mock_models/vcshandler.mock.go -package mock_models

type VCSHandler interface {
	GetVcsType() pb.SubCredType
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

	// commenting this out because it wasn't used
	//FindWebhooks iterates over existing webhooks and returns true (matches our callback urls) if one already exists
	//FindWebhooks(getWebhookURL string) bool

	//Get Repository links by account name + repo name
	GetRepoLinks(acctRepo string) (*pb.Links, error)

	// commenting out because i don't think this is used anywhere
	//GetAllCommits returns a list of commits corresponding with branch
	//GetAllCommits(acctRepo string, branch string) ([]*pb.Commit, error)

	//GetAllBranchesLastCommitData returns a list of all active branches, their last hash, and the last commit datetime
	GetAllBranchesLastCommitData(acctRepo string) ([]*pb.BranchHistory, error)

	//GetCommit will return commit information of a specific hash
	GetCommit(acctRepo, hash string) (*pb.Commit, error)

	//GetBranchLastCommitData should return the last hash and commit datetime of a specific branch
	GetBranchLastCommitData(acctRepo, branch string) (*pb.BranchHistory, error)

	//GetCommitLog will return a list of Commits, starting with the most recent and ending at the lastHash value.
	// If the lastHash commit value is never found, will return an error.
	GetCommitLog(acctRepo string, branch string, lastHash string) ([]*pb.Commit, error)

	// GetChangedFiles will get the list of files that have changed between commits. If earliestHash is not passed,
	//  then the diff list will be off of just the changed files in the latestHash. If earliesthash is passed, then it will
	//  return the changeset similar to git diff --name-only <latestHash>..<earliestHash>
	GetChangedFiles(acctRepo, latesthash, earliestHash string) ([]string, error)

	// commenting out because this method is never used
	// GetPRCommits will return a list of commits for the given url for commits. It'll call the url from (e.g. bb or github),
	//   unmarshal into its vcs-specific model, then translate to the global model to return a list of generic commits
	//GetPRCommits(url string) ([]*pb.Commit, error)

	// PostPRComment will add a comment to a pr belonging to acct/repo acctRepo and id prId with a comment that is along the lines of
	// Ocelot build has <status>.
	PostPRComment(acctRepo, prId, hash string, failed bool, buildId int64) error

	GetClient() ocenet.HttpClient
}

type Translator interface {
	//TranslatePush should take a reader body, unmarshal it to vcs-specific model, then translate it to the global Push object
	TranslatePush(reader io.Reader) (*pb.Push, error)

	//TranslatePush should take a reader body, unmarshal it to vcs-specific model, then translate it to the global PullRequest object
	TranslatePR(reader io.Reader) (*pb.PullRequest, error)
}

type CommitNotFound struct {
	hash     string
	acctRepo string
	branch   string
}

func (cnf *CommitNotFound) Error() string {
	return fmt.Sprintf("Commit hash %s was not found in the commit list for acct/repo %s at branch %s", cnf.hash, cnf.acctRepo, cnf.branch)
}

func Commit404(hash, acctRepo, branch string) *CommitNotFound {
	return &CommitNotFound{hash: hash, acctRepo: acctRepo, branch: branch}
}

type BranchNotFound struct {
	branch   string
	acctRepo string
}

func (bnf *BranchNotFound) Error() string {
	return fmt.Sprintf("Branch %s not found for acct/repo %s", bnf.branch, bnf.acctRepo)
}

func Branch404(branch, acctRepo string) *BranchNotFound {
	return &BranchNotFound{acctRepo: acctRepo, branch: branch}
}

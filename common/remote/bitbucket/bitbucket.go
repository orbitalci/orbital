package bitbucket

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/golang/protobuf/jsonpb"
	ocelog "github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"
	"github.com/shankj3/ocelot/common"
	"github.com/shankj3/ocelot/models"
	pbb "github.com/shankj3/ocelot/models/bitbucket/pb"
	"github.com/shankj3/ocelot/models/pb"
)

const DefaultCallbackURL = "http://ec2-34-212-13-136.us-west-2.compute.amazonaws.com:8088/bitbucket"
const DefaultRepoBaseURL = "https://api.bitbucket.org/2.0/repositories/%v"

//Returns VCS handler for pulling source code and auth token if exists (auth token is needed for code download)
func GetBitbucketClient(cfg *pb.VCSCreds) (models.VCSHandler, string, error) {
	bbClient := &ocenet.OAuthClient{}
	token, err := bbClient.Setup(cfg)
	if err != nil {
		return nil, "", errors.New("unable to retrieve token for " + cfg.AcctName + ".  Error: " + err.Error())
	}
	bb := GetBitbucketHandler(cfg, bbClient)
	return bb, token, nil
}

//TODO: callback url is set as env. variable on admin, or passed in via command line
//GetBitbucketHandler returns a Bitbucket handler referenced by VCSHandler interface
func GetBitbucketHandler(adminConfig *pb.VCSCreds, client ocenet.HttpClient) models.VCSHandler {
	bb := &Bitbucket{
		Client:        client,
		Marshaler:     jsonpb.Marshaler{},
		Unmarshaler:   jsonpb.Unmarshaler{AllowUnknownFields: true,},
		credConfig:    adminConfig,
		isInitialized: true,
	}
	return bb
}

//Bitbucket is a bitbucket handler responsible for finding build files and
//registering webhooks for necessary repositories
type Bitbucket struct {
	CallbackURL string
	RepoBaseURL string
	Client      ocenet.HttpClient
	Marshaler   jsonpb.Marshaler
	Unmarshaler jsonpb.Unmarshaler

	credConfig    *pb.VCSCreds
	isInitialized bool
}

//Walk iterates over all repositories and creates webhook if one doesn't
//exist. Will only work if client has been setup
func (bb *Bitbucket) Walk() error {
	if !bb.isInitialized {
		return errors.New("client has not yet been initialized, please call SetMeUp() before walking")
	}
	return bb.recurseOverRepos(fmt.Sprintf(bb.GetBaseURL(), bb.credConfig.AcctName))
}

// Get File in repo at a certain commit.
// filepath: string filepath relative to root of repo
// fullRepoName: string account_name/repo_name as it is returned in the Bitbucket api Repo Source `full_name`
// commitHash: string git hash for revision number
func (bb *Bitbucket) GetFile(filePath string, fullRepoName string, commitHash string) (bytez []byte, err error) {
	ocelog.Log().Debug("inside GetFile")
	path := fmt.Sprintf("%s/src/%s/%s", fullRepoName, commitHash, filePath)
	bytez, err = bb.Client.GetUrlRawData(fmt.Sprintf(bb.GetBaseURL(), path))
	if err != nil {
		ocelog.IncludeErrField(err).Error()
		return
	}
	return
}

//GetAllCommits /2.0/repositories/{username}/{repo_slug}/commits
func (bb *Bitbucket) GetAllCommits(acctRepo string, branch string) (*pbb.Commits, error) {
	commits := &pbb.Commits{}
	err := bb.Client.GetUrl(fmt.Sprintf(bb.GetBaseURL(), acctRepo)+"/commits/"+branch, commits)
	return commits, err
}

//GetCommitLog will return a list of Commits, starting with the most recent and ending at the lastHash value.
// If the lastHash commit value is never found, will return an error.
func (bb *Bitbucket) GetCommitLog(acctRepo, branch, lastHash string) ([]*pb.Commit, error) {
	var commits []*pb.Commit
	if lastHash == "" {
		return commits, nil
	}
	var foundLast bool
	url := fmt.Sprintf(bb.GetBaseURL(), acctRepo)+"/commits/"+branch
	for {
		if url == "" {
			break
		}
		commitz := &pbb.Commits{}
		err := bb.Client.GetUrl(url, commitz)
		if err != nil {
			return commits, err
		}
		for _, commit := range commitz.Values {
			commits = append(commits, &pb.Commit{Hash:commit.Hash, Message:commit.Message, Date:commit.Date})
			if commit.Hash == lastHash {
				foundLast = true
				break
			}
		}
		url = commitz.GetNext()
	}
	var err error
	if !foundLast {
		err = models.Commit404(lastHash, acctRepo, branch)
	}
	return commits, err
}

func (bb *Bitbucket) GetRepoDetail(acctRepo string) (pbb.PaginatedRepository_RepositoryValues, error) {
	repoVal := &pbb.PaginatedRepository_RepositoryValues{}
	err := bb.Client.GetUrl(fmt.Sprintf(DefaultRepoBaseURL, acctRepo), repoVal)
	if err != nil {
		return *repoVal, err
	}
	return *repoVal, nil
}

func (bb *Bitbucket) GetBranchLastCommitData(acctRepo, branch string) (hist *pb.BranchHistory, err error) {
	path := fmt.Sprintf("%s/refs/branches/%s", acctRepo, branch)
	url := fmt.Sprintf(bb.GetBaseURL(), path)
	var resp *http.Response
	resp, err = bb.Client.GetUrlResponse(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// status code handling using bitbucket API spec
    //   https://developer.atlassian.com/bitbucket/api/2/reference/resource/repositories/%7Busername%7D/%7Brepo_slug%7D/refs/branches/%7Bname%7D
	switch resp.StatusCode {
	case http.StatusNotFound:
		err = errors.New(fmt.Sprintf("Specified branch %s does not exist", branch))
	case http.StatusForbidden:
		err = errors.New(fmt.Sprintf("Repo %s (with branch %s) is private and these credentials are not authorized for access", acctRepo, branch))
	case http.StatusOK:
		bbBranch := &pbb.Branch{}
		reader := bufio.NewReader(resp.Body)
		if err = bb.Unmarshaler.Unmarshal(reader, bbBranch); err != nil {
			ocelog.IncludeErrField(err).Error("failed to parse response from ", url)
			return
		}
		hist = &pb.BranchHistory{Branch: branch, Hash: bbBranch.GetTarget().GetHash(), LastCommitTime: bbBranch.GetTarget().GetDate()}
		err = nil
	}
	return
}

func (bb *Bitbucket) GetAllBranchesLastCommitData(acctRepo string) ([]*pb.BranchHistory, error) {
	var branchHistory []*pb.BranchHistory
	var nextUrl string
	path := fmt.Sprintf("%s/refs/branches", acctRepo)
	nextUrl = fmt.Sprintf(bb.GetBaseURL(), path)
	for {
		branches := &pbb.PaginatedRepoBranches{}
		err := bb.Client.GetUrl(nextUrl, branches)
		if err != nil {
			return nil, err
		}
		for _, branch := range branches.GetValues() {
			branchHistory = append(branchHistory, &pb.BranchHistory{Branch: branch.Name, Hash: branch.Target.GetHash(), LastCommitTime: branch.Target.GetDate()})
		}
		nextUrl = branches.GetNext()
		if nextUrl == "" {
			break
		}
	}
	return branchHistory, nil
}


//CreateWebhook will create webhook at specified webhook url
func (bb *Bitbucket) CreateWebhook(webhookURL string) error {
	if !bb.FindWebhooks(webhookURL) {
		//create webhook if one does not already exist
		newWebhook := &pbb.CreateWebhook{
			Description: "marianne did this",
			Active:      true,
			Url:         bb.GetCallbackURL(),
			Events:      common.BitbucketEvents,
		}
		webhookStr, err := bb.Marshaler.MarshalToString(newWebhook)
		if err != nil {
			ocelog.IncludeErrField(err).Fatal("failed to convert webhook to json string")
			return err
		}
		err = bb.Client.PostUrl(webhookURL, webhookStr, nil)
		if err != nil {
			return err
		}
		ocelog.Log().Debug("subscribed to webhook for ", webhookURL)
	}
	return nil
}

//GetCallbackURL is a getter for retrieving callbackURL for bitbucket webhooks
func (bb *Bitbucket) GetCallbackURL() string {
	if len(bb.CallbackURL) > 0 {
		return bb.CallbackURL
	}
	return DefaultCallbackURL
}

//SetCallbackURL sets callback urls to be used for webhooks
func (bb *Bitbucket) SetCallbackURL(callbackURL string) {
	bb.CallbackURL = callbackURL
}

func (bb *Bitbucket) SetBaseURL(baseURL string) {
	bb.RepoBaseURL = baseURL
}

func (bb *Bitbucket) GetBaseURL() string {
	if len(bb.RepoBaseURL) > 0 {
		return bb.RepoBaseURL
	}
	return DefaultRepoBaseURL
}

//recursively iterates over all repositories and creates webhook
func (bb *Bitbucket) recurseOverRepos(repoUrl string) error {
	if repoUrl == "" {
		return nil
	}
	repositories := &pbb.PaginatedRepository{}
	//todo: error pages from bitbucket??? these need to bubble up to client
	err := bb.Client.GetUrl(repoUrl, repositories)
	if err != nil {
		return err
	}

	for _, v := range repositories.GetValues() {
		ocelog.Log().Debug(fmt.Sprintf("found repo %v", v.GetFullName()))
		err = bb.recurseOverFiles(v.GetLinks().GetSource().GetHref(), v.GetLinks().GetHooks().GetHref())
		if err != nil {
			return err
		}
	}
	return bb.recurseOverRepos(repositories.GetNext())
}

//recursively iterates over all source files trying to find build file
func (bb Bitbucket) recurseOverFiles(sourceFileUrl string, webhookUrl string) error {
	if sourceFileUrl == "" {
		return nil
	}
	repositories := &pbb.PaginatedRootDirs{}
	err := bb.Client.GetUrl(sourceFileUrl, repositories)
	if err != nil {
		return err
	}
	for _, v := range repositories.GetValues() {
		if v.GetType() == "commit_file" && len(v.GetAttributes()) == 0 && v.GetPath() == common.BuildFileName {
			ocelog.Log().Debug("holy crap we actually an ocelot.yml file")
			err = bb.CreateWebhook(webhookUrl)
			if err != nil {
				return err
			}
		}
	}
	return bb.recurseOverFiles(repositories.GetNext(), webhookUrl)
}

//recursively iterates over all webhooks and returns true (matches our callback urls) if one already exists
func (bb *Bitbucket) FindWebhooks(getWebhookURL string) bool {
	if getWebhookURL == "" {
		return false
	}
	webhooks := &pbb.GetWebhooks{}
	bb.Client.GetUrl(getWebhookURL, webhooks)

	for _, wh := range webhooks.GetValues() {
		if wh.GetUrl() == bb.GetCallbackURL() {
			return true
		}
	}

	return bb.FindWebhooks(webhooks.GetNext())
}


func (bb *Bitbucket) TranslatePush(reader io.Reader) (*pb.Push, error) {
	push := &pbb.RepoPush{}
	err := bb.Unmarshaler.Unmarshal(reader, push)
	if err != nil {
		return nil, err
	}

	if len(push.Push.Changes) < 1 {
		ocelog.Log().Error("no commits found in push")
		// todo: evaluate if this is actually what should happen, what about pushing tags?
		return nil, errors.New("no commits found in push")
	}
	if len(push.Push.Changes) > 1 {
		ocelog.Log().Errorf("length of push changes is > 1, changes are %#v", push.Push.Changes)
		return nil, errors.New("too many changesets")
	}
	var commits []*pb.Commit
	changeset := push.Push.Changes[0]
	for _, commit := range changeset.Commits {
		commits = append(commits, &pb.Commit{Hash:commit.Hash, Date:commit.Date, Message:commit.Message, Author:&pb.User{UserName:commit.Author.Username}})
	}
	translPush := &pb.Push{
		Repo: &pb.Repo{Name: push.Repository.FullName, AcctRepo:push.Repository.FullName, RepoLink: push.Repository.Links.Html.Href},
		User: &pb.User{UserName: push.Actor.Username},
		Commits: commits,
		HeadCommit: &pb.Commit{Hash:changeset.New.Target.Hash, Message:changeset.New.Target.Hash, Date:changeset.New.Target.Date, Author:&pb.User{UserName:changeset.New.Target.Author.Username}},
	}
	return translPush, nil
}

func (bb *Bitbucket) TranslatePR(reader io.Reader) (*pb.PullRequest, error) {
	pr := &pbb.PullRequest{}
	err := bb.Unmarshaler.Unmarshal(reader, pr)
	if err != nil {
		return nil, err
	}
	prN := &pb.PullRequest{
		Description: pr.Pullrequest.Description,
		Urls: &pb.PrUrls{
			Commits: pr.Pullrequest.Links.Commits.Href,
			Comments: pr.Pullrequest.Links.Comments.Href,
			Statuses: pr.Pullrequest.Links.Statuses.Href,
		},
		Title: pr.Pullrequest.Title,
		Repo: &pb.Repo{
				Name: pr.Pullrequest.Source.Repository.FullName,
				AcctRepo: pr.Pullrequest.Source.Repository.FullName,
				RepoLink: pr.Pullrequest.Source.Repository.Links.Html.Href,
		},
		DestRepo: &pb.Repo{
			Name: pr.Pullrequest.Destination.Repository.FullName,
			AcctRepo: pr.Pullrequest.Destination.Repository.FullName,
			RepoLink: pr.Pullrequest.Destination.Repository.Links.Html.Href,
		},
	}
	return prN, nil
}

func (bb *Bitbucket) GetPRCommits(url string) ([]*pb.Commit, error) {
	var commits []*pb.Commit
	for {
		if url == "" {
			break
		}
		commitz := &pbb.Commits{}
		err := bb.Client.GetUrl(url, commitz)
		if err != nil {
			return commits, err
		}
		for _, commit := range commitz.Values {
			commits = append(commits, &pb.Commit{Hash:commit.Hash, Message:commit.Message, Date:commit.Date, Author:&pb.User{UserName: commit.Author.Username}})
		}
		url = commitz.GetNext()
	}
	return commits, nil
}
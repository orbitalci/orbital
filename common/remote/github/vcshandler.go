package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/golang/protobuf/jsonpb"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"

	ocelog "github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"
	"github.com/shankj3/ocelot/common"
	"github.com/shankj3/ocelot/models"
	gpb "github.com/shankj3/ocelot/models/github/pb"
	"github.com/shankj3/ocelot/models/pb"
)

const DefaultCallbackURL = "http://ec2-34-212-13-136.us-west-2.compute.amazonaws.com:8088/github"

//Returns VCS handler for pulling source code and auth token if exists (auth token is needed for code download)
func GetGithubClient(creds *pb.VCSCreds) (models.VCSHandler, string, error) {
	client := &ocenet.OAuthClient{}
	token, err := client.SetupStaticToken(creds)
	if err != nil {
		return nil, "", errors.New("unable to retrieve token for " + creds.AcctName + ".  Error: " + err.Error())
	}
	gh := GetGithubHandler(creds, client)
	return gh, token, nil
}

func GetGithubHandler(cred *pb.VCSCreds, cli ocenet.HttpClient) *githubVCS {
	return &githubVCS{
		Client:        cli,
		Marshaler:     jsonpb.Marshaler{},
		Unmarshaler:   jsonpb.Unmarshaler{AllowUnknownFields: true},
		credConfig:    cred,
		ghClient:      github.NewClient(cli.GetAuthClient()),
		ctx:           context.Background(),
	}
}

type githubVCS struct {
	CallbackURL   string
	RepoBaseURL   string
	Client 		  ocenet.HttpClient
	ghClient      *github.Client
	ctx           context.Context
	Marshaler     jsonpb.Marshaler
	Unmarshaler   jsonpb.Unmarshaler
	credConfig    *pb.VCSCreds
	baseUrl       string
	// for testing
	setCommentId  int64
}

func (gh *githubVCS) GetCallbackURL() string {
	if gh.CallbackURL == "" {
		return DefaultCallbackURL
	}
	return gh.CallbackURL
}

func (gh *githubVCS) SetCallbackURL(cbUrl string) {
	gh.CallbackURL = cbUrl
}

func (gh *githubVCS) GetClient() ocenet.HttpClient {
	return gh.Client
}

func (gh *githubVCS) GetBaseURL() string {
	if gh.baseUrl != "" {
		return gh.baseUrl
	}
	return "https://api.github.com/%s"
}

func (gh *githubVCS) SetBaseURL(baseUrl string) {
	gh.baseUrl = baseUrl
}

//Walk iterates over all repositories and creates webhook if one doesn't
//exist. Will only work if client has been setup
func (gh *githubVCS) Walk() error {
	return gh.recurseOverRepos(0)
}

// recurseOverRepos will iterate over every repository checking for an ocelot.yml
// if one is found, then a webhook will be created
func (gh *githubVCS) recurseOverRepos(pageNum int) error {

	opts := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 50, Page: pageNum},
	}
	repos, resp, err := gh.ghClient.Repositories.List(gh.ctx, "", opts)
	if err != nil {
		return errors.Wrap(err, "unable to list all repos")
	}
	resp.Body.Close()
	for _, repo := range repos {
		statusCode, erro := gh.checkForOcelotFile(repo.GetContentsURL())
		if erro != nil {
			return erro
		}
		if statusCode == http.StatusOK {
			if err = gh.CreateWebhook(repo.GetHooksURL()); err != nil {
				return errors.Wrap(err, "unable to create webhook")
			}
		}
	}
	if resp.NextPage == 0 {
		return nil
	}
	return gh.recurseOverRepos(resp.NextPage)
}

func (gh *githubVCS) checkForOcelotFile(contentsUrl string) (int, error) {
	resp, err := gh.Client.GetUrlResponse(getUrlForFileFromContentsUrl(contentsUrl, common.BuildFileName))
	if err != nil {
		return 0, errors.Wrap(err, "unable to see if ocelot.yml exists")
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func (gh *githubVCS) CreateWebhook(hookUrl string) error {
	// create it, if it already exists it'll return a 422
	hookReq := &gpb.Hook{
		Active: true,
		Events: []string{"push", "pull_request"},
		Config: &gpb.Hook_Config{Url: DefaultCallbackURL, ContentType: "json"},
	}
	bits, err := json.Marshal(hookReq)
	if err != nil {
		return errors.Wrap(err, "couldn't marshal hook request to json")
	}
	var resp *http.Response
	resp, err = gh.Client.GetAuthClient().Post(hookUrl, "application/json", bytes.NewReader(bits))
	if err != nil {
		failedGHRemoteCalls.WithLabelValues("CreateWebhook").Inc()
		return errors.Wrap(err, "unable to complete webhook create")
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	ghErr := &gpb.Error{}
	_ = jsonpb.Unmarshal(resp.Body, ghErr)
	if resp.StatusCode == http.StatusUnprocessableEntity {
		if ghErr.Message == "Validation Failed" {
			return nil
		}
	}
	return errors.New(resp.Status +": "+ ghErr.Message)
}


func (gh *githubVCS) GetFile(filePath string, fullRepoName string, commitHash string) (bytez []byte, err error) {
	acct, repo := splitAcctRepo(fullRepoName)
	getOpts := &github.RepositoryContentGetOptions{Ref: commitHash}
	var contents io.ReadCloser
	contents, err = gh.ghClient.Repositories.DownloadContents(gh.ctx, acct, repo, filePath, getOpts)
	if err != nil {
		failedGHRemoteCalls.WithLabelValues("GetFile").Inc()
		ocelog.IncludeErrField(err).Error("cannot get file contentx")
		err = errors.Wrap(err, "unable to get file contents")
		return
	}
	defer contents.Close()
	bytez, err = ioutil.ReadAll(contents)
	return
}

func (gh *githubVCS) GetRepoLinks(acctRepo string) (*pb.Links, error) {
	acct, repo := splitAcctRepo(acctRepo)
	repository, resp, err := gh.ghClient.Repositories.Get(gh.ctx, acct, repo)
	if err != nil {
		failedGHRemoteCalls.WithLabelValues("GetRepoLinks").Inc()
		ocelog.IncludeErrField(err).Error("cannot get repo links")
		return nil, errors.Wrap(err, "unable to get repository links")
	}
	defer resp.Body.Close()
	links := &pb.Links{
		Commits: repository.GetCommitsURL(),
		Branches: repository.GetBranchesURL(),
		Tags: repository.GetTagsURL(),
		Hooks: repository.GetHooksURL(),
		Pullrequests: repository.GetPullsURL(),
	}
	return links, nil
}

func (gh *githubVCS) GetAllBranchesLastCommitData(acctRepo string) ([]*pb.BranchHistory, error) {
	var branchesHistory []*pb.BranchHistory
	acct, repo := splitAcctRepo(acctRepo)
	opts := &github.ListOptions{PerPage: 50}
	for {
		branches, resp, err := gh.ghClient.Repositories.ListBranches(gh.ctx, acct, repo, opts)
		if err != nil {
			failedGHRemoteCalls.WithLabelValues("GetAllBrancheseLastCommitData").Inc()
			ocelog.IncludeErrField(err).WithField("acctRepo", acctRepo).Error("cannot get branch data")
			return nil, err
		}
		for _, branch := range branches {
			branchesHistory = append(branchesHistory, translateToBranchHistory(branch))
		}
		resp.Body.Close()
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return branchesHistory, nil
}


func (gh *githubVCS) GetBranchLastCommitData(acctRepo, branch string) (history *pb.BranchHistory, err error) {
	acct, repo := splitAcctRepo(acctRepo)
	brch, resp, err := gh.ghClient.Repositories.GetBranch(gh.ctx, acct, repo, branch)
	if err != nil {
		ocelog.IncludeErrField(err).Error("unable to get last commit data")
		failedGHRemoteCalls.WithLabelValues("GetBranchLastCommitData").Inc()
		err = errors.Wrap(err, "unable to get last commit data")
		return
	}
	defer resp.Body.Close()
	history = translateToBranchHistory(brch)
	return
}

func (gh *githubVCS) GetCommitLog(acctRepo string, branch string, lastHash string) (commits []*pb.Commit, err error) {
	acct, repo := splitAcctRepo(acctRepo)
	opt := &github.CommitsListOptions{
		SHA: branch,
		ListOptions: github.ListOptions{
			PerPage: 40,
		},
	}
	for {
		ghCommits, resp, err := gh.ghClient.Repositories.ListCommits(gh.ctx, acct, repo, opt)
		if err != nil {
			failedGHRemoteCalls.WithLabelValues("GetCommitLog").Inc()
			ocelog.IncludeErrField(err).Error("unable to get list of commits!")
			return nil, errors.Wrap(err, "unable to get list of commits")
		}
		resp.Body.Close()
		for _, ghCommit := range ghCommits {
			commits = append(commits, translateToCommit(ghCommit))
			if ghCommit.GetSHA() == lastHash {
				goto RETURN
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
RETURN:
	return
}

func (gh *githubVCS) PostPRComment(acctRepo, prId, hash string, failed bool, buildId int64) error {
	acct, repo := splitAcctRepo(acctRepo)
	var status string
	switch failed {
	case true:
		status = "FAILED"
	case false:
		status = "PASSED"
	}
	content := fmt.Sprintf("Ocelot build has **%s** for commit **%s**.\n\nRun `ocelot status -build-id %d` for detailed stage status, and `ocelot run -build-id %d` for complete build logs.", status, hash, buildId, buildId)
	prIdInt, err := strconv.Atoi(prId)
	if err != nil {
		return errors.Wrap(err, "invalid pr id")
	}
	comment := &github.IssueComment{Body: &content}
	cmt, resp, err := gh.ghClient.Issues.CreateComment(gh.ctx, acct, repo, prIdInt, comment)
	if err != nil {
		failedGHRemoteCalls.WithLabelValues("PostPRComment").Inc()
		ocelog.IncludeErrField(err).WithField("prId", prId).WithField("buildId", buildId).Error("unable to create pr comment")
		return errors.Wrap(err, "unable to create a pr comment")
	}
	resp.Body.Close()
	gh.setCommentId = cmt.GetID()
	return nil
}

// for testing
func (gh *githubVCS) deleteIssueComment(account, repo string, commentId int64) error {
	resp, err := gh.ghClient.Issues.DeleteComment(gh.ctx, account, repo, commentId)
	if err != nil {
		ocelog.IncludeErrField(err).Error("bad delete")
		return err
	}
	resp.Body.Close()
	return nil
}

// for testing
func (gh *githubVCS) getIssueComment(account, repo string, commentID int64) error {
	comment, resp, err := gh.ghClient.Issues.GetComment(gh.ctx, account, repo, commentID)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if comment == nil {
		return errors.New("not found")
	}
	return nil
}

func (gh *githubVCS) GetChangedFiles(acctRepo, latesthash, earliestHash string) ([]string, error) {
	var changedFiles []string
	//GET /repos/:owner/:repo/compare/:base...:head
	acct, repo := splitAcctRepo(acctRepo)
	compare, resp, err := gh.ghClient.Repositories.CompareCommits(gh.ctx, acct, repo, earliestHash, latesthash)
	if err != nil {
		failedGHRemoteCalls.WithLabelValues("GetChangedFiles").Inc()
		ocelog.IncludeErrField(err).WithField("latestHash", latesthash).WithField("earliestHash", earliestHash).Error("unable to get changed files")
		return nil, errors.Wrap(err, "unable to get changed files")
	}
	resp.Body.Close()
	for _, file := range compare.Files {
		changedFiles = append(changedFiles, file.GetFilename())
	}
	return changedFiles, nil
}

func (gh *githubVCS) GetCommit(acctRepo, hash string) (*pb.Commit, error) {
	acct, repo := splitAcctRepo(acctRepo)
	commit, resp, err := gh.ghClient.Repositories.GetCommit(gh.ctx, acct, repo, hash)
	if err != nil {
		failedGHRemoteCalls.WithLabelValues("GetCommit").Inc()
		ocelog.IncludeErrField(err).Error("cannot get commit")
		return nil, errors.Wrap(err, "cannot get commit")
	}
	resp.Body.Close()
	return translateToCommit(commit), nil
}

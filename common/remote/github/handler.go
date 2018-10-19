package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"github.com/tomnomnom/linkheader"

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
	token, err := client.Setup(creds)
	if err != nil {
		return nil, "", errors.New("unable to retrieve token for " + creds.AcctName + ".  Error: " + err.Error())
	}
	gh := GetGithubHandler(creds, client)
	return gh, token, nil
}

func GetGithubHandler(cred *pb.VCSCreds, cli ocenet.HttpClient) *github {
	return &github{
		Client:        cli,
		Marshaler:     jsonpb.Marshaler{},
		Unmarshaler:   jsonpb.Unmarshaler{AllowUnknownFields: true},
		credConfig:    cred,
	}
}

type github struct {
	CallbackURL   string
	RepoBaseURL   string
	Client 		  ocenet.HttpClient
	Marshaler     jsonpb.Marshaler
	Unmarshaler   jsonpb.Unmarshaler
	credConfig    *pb.VCSCreds
	baseUrl       string
}

func (gh *github) GetCallbackURL() string {
	if gh.CallbackURL == "" {
		return DefaultCallbackURL
	}
	return gh.CallbackURL
}

func (gh *github) SetCallbackURL(cbUrl string) {
	gh.CallbackURL = cbUrl
}

func (gh *github) GetClient() ocenet.HttpClient {
	return gh.Client
}

func (gh *github) GetBaseURL() string {
	if gh.baseUrl != "" {
		return gh.baseUrl
	}
	return "https://api.github.com/%s"
}

func (gh *github) SetBaseURL(baseUrl string) {
	gh.baseUrl = baseUrl
}

//Walk iterates over all repositories and creates webhook if one doesn't
//exist. Will only work if client has been setup
func (gh *github) Walk() error {
	return gh.recurseOverRepos(fmt.Sprintf(gh.GetBaseURL(), ALLREPOS))
}

func getNextPage(headers http.Header) string {
	link := headers.Get("Link")
	if link == "" {
		return ""
	}
	links := linkheader.Parse(link)
	nexts := links.FilterByRel("next")
	if len(nexts) == 0 {
		return ""
	}
	return nexts[0].URL
}

// recurseOverRepos will iterate over every repository checking for an ocelot.yml
// if one is found, then a webhook will be created
func (gh *github) recurseOverRepos(repoUrl string) error {
	if repoUrl == "" {
		return nil
	}
	var repos []*gpb.Repository
	resp, err := gh.Client.GetUrlResponse(repoUrl)
	if err != nil {
		return errors.Wrap(err, "unable to get list of repositories to check for webhooks")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "unable to read url response")
	}
	if err = json.Unmarshal(body, repos); err != nil {
		return errors.Wrap(err, "unable to unmarshal to repo list")
	}
	for _, repo := range repos {
		statusCode, erro := gh.checkForOcelotFile(repo.ContentsUrl)
		if erro != nil {
			return erro
		}
		if statusCode == http.StatusOK {
			if err = gh.CreateWebhook(repo.HooksUrl); err != nil {
				return errors.Wrap(err, "unable to create webhook")
			}
		}
	}
	return gh.recurseOverRepos(getNextPage(resp.Header))
}

func (gh *github) checkForOcelotFile(contentsUrl string) (int, error) {
	resp, err := gh.Client.GetUrlResponse(getUrlForFileFromContentsUrl(contentsUrl, common.BuildFileName))
	if err != nil {
		return 0, errors.Wrap(err, "unable to see if ocelot.yml exists")
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func (gh *github) CreateWebhook(hookUrl string) error {
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
	var ghErr *gpb.Error
	_ = jsonpb.Unmarshal(resp.Body, ghErr)
	if resp.StatusCode == http.StatusUnprocessableEntity {
		if ghErr.Message == "Validation Failed" {
			return nil
		}
	}
	return errors.New(resp.Status +": "+ ghErr.Message)
}

func (gh *github) GetFile(filePath string, fullRepoName string, commitHash string) (bytez []byte, err error) {
	url := fmt.Sprintf(gh.GetBaseURL(), buildFilePath(fullRepoName, commitHash))
	bytez, err = gh.Client.GetUrlRawData(url)
	if err != nil {
		failedGHRemoteCalls.WithLabelValues("GetFile").Inc()
		ocelog.IncludeErrField(err).Error("unable to get file")
	}
	return
}

func (gh *github) GetRepoLinks(acctRepo string) (*pb.Links, error) {
	url := fmt.Sprintf(gh.GetBaseURL(), buildRepoPath(acctRepo))
	repo := &gpb.Repository{}
	err := gh.Client.GetUrl(url, repo)
	if err != nil {
		return nil, err
	}
	links := &pb.Links{
		Commits: repo.CommitsUrl,
		Branches: repo.BranchesUrl,
		Tags: repo.TagsUrl,
		Hooks: repo.HooksUrl,
		Pullrequests: repo.PullsUrl,
	}
	return links, nil
}

func (gh *github) GetAllBranchesLastCommitData(acctRepo string) ([]*pb.BranchHistory, error) {
	var branchesHistory []*pb.BranchHistory
	url := fmt.Sprintf(gh.GetBaseURL(), buildBranchesPath(acctRepo, ""))
	resp, err := gh.Client.GetUrlResponse(url)
	if err != nil {
		failedGHRemoteCalls.WithLabelValues("GetAllBranchesLastCommitData").Inc()
		return nil, err
	}
	defer resp.Body.Close()
	bytez, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusOK {
		branchesHistory, err = gh.buildBranchHistories(bytez)
		if err != nil {
			return nil, err
		}
	} else {
		var ghErr *gpb.Error
		err = json.Unmarshal(bytez, ghErr)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(fmt.Sprintf("github returned unexpected error: %s", ghErr.Message))
	}
	return branchesHistory, nil
}

func (gh *github) buildBranchHistories(responseBytes []byte) (histories []*pb.BranchHistory, err error) {
	var branchesLastCommit []*gpb.BranchLastCommit
	err = json.Unmarshal(responseBytes, branchesLastCommit)
	if err != nil {
		return
	}
	for _, branch := range branchesLastCommit {
		stamp, err := gh.getCommitTime(branch.Commit.Url)
		if err != nil {
			return
		}
		histories = append(histories, &pb.BranchHistory{Hash: branch.Commit.Sha, Branch: branch.Name, LastCommitTime: stamp})
	}
	return
}

func (gh *github) getCommitTime(commitUrl string) (*timestamp.Timestamp, error) {
	commit := &gpb.Commit{}
	err := gh.Client.GetUrl(commitUrl, commit)
	if err != nil {
		failedGHRemoteCalls.WithLabelValues("getCommitTime").Inc()
		return nil, err
	}
	return commit.Commit.Commiter.Date, err
}

func (gh *github) GetBranchLastCommitData(acctRepo, branch string) (history *pb.BranchHistory, err error) {
	url := fmt.Sprintf(gh.GetBaseURL(), buildBranchesPath(acctRepo, ""))
	var resp *http.Response
	resp, err = gh.Client.GetUrlResponse(url)
	if err != nil {
		failedGHRemoteCalls.WithLabelValues("GetBranchLastCommitData").Inc()
		return nil, err
	}
	defer resp.Body.Close()
	var responseBody []byte
	responseBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusOK {
		var branchLastCommit *gpb.BranchLastCommit
		err = json.Unmarshal(responseBody, branchLastCommit)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal response body to BranchLastCommit")
		}
		history.LastCommitTime = branchLastCommit.Commit.Commit.Commiter.Date
		history.Branch = branchLastCommit.Name
		history.Hash = branchLastCommit.Commit.Sha
	} else {
		var ghErr *gpb.Error
		err = json.Unmarshal(responseBody, ghErr)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal body to github error")
		}
		err = errors.New(fmt.Sprintf("unexpected error while getting last commit data: %s", ghErr.Message))
	}
	return
}

func (gh *github) GetCommitLog(acctRepo string, branch string, lastHash string) (commits []*pb.Commit, err error) {
	// todo : add to go-til/net query params
	//https://stackoverflow.com/questions/30652577/go-doing-a-get-request-and-building-the-querystring/30657518
	path := buildCommitsPath(acctRepo) + "?sha=" + branch
	url := fmt.Sprintf(gh.GetBaseURL(), path)
	for {
		if url == "" || err != nil {
			break
		}
		var pagedCommits []*pb.Commit
		pagedCommits, url, err = gh.getCommitsList(url, lastHash)
		commits = append(commits, pagedCommits...)

	}
	// if there are no errors at this point, then it means the entire branch history was exhausted without finding the hash
	if err == nil {
		err = models.Commit404(lastHash, acctRepo, branch)
	}
	// if HitTheCommit error was returned, it means it successfully found lastHash and the commit list is complete. no need to error here
	if err == HitTheCommit {
		err = nil
	}
	return
}

var HitTheCommit = errors.New("hit the last commit that was supplied")

// getCommitsList will use the url provided to request a list of commits, attempt to unmarshal to a list of
//  github commit structs, then translate that commit into a generic commit. if in iterating over the commit it
//  reaches a commit with the same sha as lastCommitHash, then it will return a HitTheCommit error which should be
//  handled by the calling function. it will also parse the headers for a Link and return the next url found in
//  that Link if it is found.
func (gh *github) getCommitsList(url, lastCommitHash string) (commits []*pb.Commit, nextUrl string, err error) {
	var resp *http.Response
	resp, err = gh.Client.GetUrlResponse(url)
	if err != nil {
		failedGHRemoteCalls.WithLabelValues("getCommitsList").Inc()
		return
	}
	defer resp.Body.Close()
	var responseBody []byte
	responseBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode == http.StatusOK {
		var ghCommits []*gpb.Commit
		err = json.Unmarshal(responseBody, ghCommits)
		if err != nil {
			return nil,"", errors.Wrap(err, "unable to unmarshal response body to list of Commits")
		}
		// look at all the commits, translate them to *pb.Commit, and check for the lastCommitHash
		for _, commit := range ghCommits {
			commits = append(commits, &pb.Commit{Hash: commit.Sha, Message: commit.Commit.Message, Date: commit.Commit.Commiter.Date, Author: &pb.User{UserName: commit.Commit.Commiter.Name}})
			if strings.Contains(commit.Sha, lastCommitHash) {
				return nil, "", HitTheCommit
			}
		}
	} else {
		var ghErr *gpb.Error
		if err = json.Unmarshal(responseBody, ghErr); err != nil {
			return nil, "", errors.Wrap(err, "unable to unmarshal response body to github error obj")
		}
		return nil, "", errors.New("unable to retrieve list of commits, error is " + ghErr.Message)
 	}
 	nextUrl = getNextPage(resp.Header)
 	return
}


func (gh *github) GetPRCommits(url string) (commits []*pb.Commit, err error) {
	var pagedCommits []*pb.Commit
	for {
		if url == "" || err != nil {
			break
		}
		// we want to just get all commits under the pr commits list
		pagedCommits, url, err = gh.getCommitsList(url, "-random-ocelot&&&string")
		commits = append(commits, pagedCommits...)
	}
	return
}

func (gh *github) PostPRComment(acctRepo, prId, hash string, failed bool, buildId int64) error {
	url := fmt.Sprintf(gh.GetBaseURL(), buildPrCommentsPath(acctRepo, prId))
	var status string
	switch failed {
	case true:
		status = "FAILED"
	case false:
		status = "PASSED"
	}
	content := fmt.Sprintf("Ocelot build has **%s** for commit **%s**.\n\nRun `ocelot status -build-id %d` for detailed stage status, and `ocelot run -build-id %d` for complete build logs.", status, hash, buildId, buildId)
	body := map[string]string{
		"body": content,
	}
	bits, err := json.Marshal(body)
	if err != nil {
		return errors.Wrap(err, "unable to marshal pr comment body to json")
	}
	resp, err := gh.Client.GetAuthClient().Post(url, "application/json", bytes.NewReader(bits))
	if err != nil {
		return errors.Wrap(err, "unable to post pr comment to github")
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return nil
	} else {
		var ghErr *gpb.Error
		bytz, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "reader readall error")
		}
		if err = json.Unmarshal(bytz, ghErr); err != nil {
			return errors.Wrap(err, "unable to unmarshal to github error")
		}
		return errors.New("error posting PR comment: " + ghErr.Message)
	}
}
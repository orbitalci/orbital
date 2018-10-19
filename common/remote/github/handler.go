package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/protobuf/jsonpb"
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

	models.VCSHandler
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
	return "https://api.github.com/%s"
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
		resp, err = gh.Client.GetUrlResponse(getUrlForFileFromContentsUrl(repo.ContentsUrl, common.BuildFileName))
		if err != nil {
			return errors.Wrap(err, "unable to see if ocelot.yml exists")
		}
		if resp.StatusCode == http.StatusOK {
			if err = gh.CreateWebhook(repo.HooksUrl); err != nil {
				return errors.Wrap(err, "unable to create webhook")
			}
		}
	}
	return gh.recurseOverRepos(getNextPage(resp.Header))
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
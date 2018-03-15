package handler

import (
	ocelog "bitbucket.org/level11consulting/go-til/log"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"errors"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
)

const DefaultCallbackURL = "http://ec2-34-212-13-136.us-west-2.compute.amazonaws.com:8088/bitbucket"
const DefaultRepoBaseURL = "https://api.bitbucket.org/2.0/repositories/%v"

const ChangeSetRepoBaseURL = "https://api.bitbucket.org/1.0/repositories/%v/changesets/%v"

//TODO: callback url is set as env. variable on admin, or passed in via command line
//GetBitbucketHandler returns a Bitbucket handler referenced by VCSHandler interface
func GetBitbucketHandler(adminConfig *models.VCSCreds, client ocenet.HttpClient) VCSHandler {
	bb := &Bitbucket{
		Client: client,
		Marshaler: jsonpb.Marshaler{},
		credConfig: adminConfig,
		isInitialized: true,
	}
	return bb
}

//Bitbucket is a bitbucket handler responsible for finding build files and
//registering webhooks for necessary repositories
type Bitbucket struct {
	CallbackURL string
	RepoBaseURL	string
	Client      ocenet.HttpClient
	Marshaler   jsonpb.Marshaler

	credConfig    *models.VCSCreds
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
		ocelog.IncludeErrField(err)
		return
	}
	return
}

func (bb *Bitbucket) GetRepoDetail(acctRepo string) (pb.PaginatedRepository_RepositoryValues, error){
	repoVal := &pb.PaginatedRepository_RepositoryValues{}
	err := bb.Client.GetUrl(fmt.Sprintf(DefaultRepoBaseURL, acctRepo), repoVal)
	if err != nil {
		return *repoVal, err
	}
	return *repoVal, nil
}

//CreateWebhook will create webhook at specified webhook url
func (bb *Bitbucket) CreateWebhook(webhookURL string) error {
	if !bb.FindWebhooks(webhookURL) {
		//create webhook if one does not already exist
		newWebhook := &pb.CreateWebhook{
			Description: "marianne did this",
			Active:      true,
			Url:         bb.GetCallbackURL(),
			Events:      models.BitbucketEvents,
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
	if len (bb.RepoBaseURL) > 0 {
		return bb.RepoBaseURL
	}
	return DefaultRepoBaseURL
}

func (bb *Bitbucket) GetHashDetail(acctRepo, hash string) (pb.ChangeSetV1, error) {
	if len(acctRepo) == 0 || len(hash) == 0 {
		return pb.ChangeSetV1{}, errors.New("please pass a valid acct/repo and hash")
	}
	hashDetail := &pb.ChangeSetV1{}
	err := bb.Client.GetUrl(fmt.Sprintf(ChangeSetRepoBaseURL, acctRepo, hash), hashDetail)
	if err != nil {
		ocelog.IncludeErrField(err)
		return *hashDetail, err
	}

	return *hashDetail, nil
}

//recursively iterates over all repositories and creates webhook
func (bb *Bitbucket) recurseOverRepos(repoUrl string) error {
	if repoUrl == "" {
		return nil
	}
	repositories := &pb.PaginatedRepository{}
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
	repositories := &pb.PaginatedRootDirs{}
	err := bb.Client.GetUrl(sourceFileUrl, repositories)
	if err != nil {
		return err
	}
	for _, v := range repositories.GetValues() {
		if v.GetType() == "commit_file" && len(v.GetAttributes()) == 0 && v.GetPath() == models.BuildFileName {
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
	webhooks := &pb.GetWebhooks{}
	bb.Client.GetUrl(getWebhookURL, webhooks)

	for _, wh := range webhooks.GetValues() {
		if wh.GetUrl() == bb.GetCallbackURL() {
			return true
		}
	}

	return bb.FindWebhooks(webhooks.GetNext())
}

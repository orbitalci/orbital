package handler

//TODO: add interface once we have more than just bitbucket
import (
	ocelog "bitbucket.org/level11consulting/go-til/log"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"errors"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
)

const DefaultCallbackURL = "http://52.26.105.112/bitbucket"
const BitbucketRepoBase = "https://api.bitbucket.org/2.0/repositories/%v"

//Bitbucket is a bitbucket handler responsible for finding build files and
//registering webhooks for necessary repositories
type Bitbucket struct {
	CallbackURL string
	Client      ocenet.HttpClient
	Marshaler   jsonpb.Marshaler

	credConfig    *models.Credentials
	isInitialized bool
}

// Takes in admin config creds, returns any errors that may happen during setup
func (bb *Bitbucket) SetMeUp(adminConfig *models.Credentials, client ocenet.HttpClient) error {
	bb.Client = client
	bb.Marshaler = jsonpb.Marshaler{}
	bb.credConfig = adminConfig
	bb.isInitialized = true
	return nil
}

//Walk iterates over all repositories and creates webhook if one doesn't
//exist. Will only work if client has been setup
func (bb *Bitbucket) Walk() error {
	if !bb.isInitialized {
		return errors.New("client has not yet been initialized, please call SetMeUp() before walking")
	}
	return bb.recurseOverRepos(fmt.Sprintf(BitbucketRepoBase, bb.credConfig.AcctName))
}

// Get File in repo at a certain commit.
// filepath: string filepath relative to root of repo
// fullRepoName: string account_name/repo_name as it is returned in the Bitbucket api Repo Source `full_name`
// commitHash: string git hash for revision number
func (bb *Bitbucket) GetFile(filePath string, fullRepoName string, commitHash string) (bytez []byte, err error) {
	ocelog.Log().Debug("inside GetFile")
	path := fmt.Sprintf("%s/src/%s/%s", fullRepoName, commitHash, filePath)
	bytez, err = bb.Client.GetUrlRawData(fmt.Sprintf(BitbucketRepoBase, path))
	if err != nil {
		return
	}
	return
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

//recursively iterates over all repositories and creates webhook
func (bb *Bitbucket) recurseOverRepos(repoUrl string) error {
	if repoUrl == "" {
		return nil
	}
	repositories := &pb.PaginatedRepository{}
	err := bb.Client.GetUrl(repoUrl, repositories)
	if err != nil {
		return err
	}

	for _, v := range repositories.GetValues() {
		fmt.Printf("found repo %v\n", v.GetFullName())
		err = bb.CreateWebhook(v.GetLinks().GetHooks().GetHref())
		if err != nil {
			return err
		}
	}
	return bb.recurseOverRepos(repositories.GetNext())
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

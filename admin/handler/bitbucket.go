package handler

//TODO: add interface once we have more than just bitbucket
import (
	"context"
	//"errors"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/shankj3/ocelot/admin/models"
	"github.com/shankj3/ocelot/util/ocelog"
	"github.com/shankj3/ocelot/util/ocenet"
	pb "github.com/shankj3/ocelot/protos/out"
	"golang.org/x/oauth2/clientcredentials"
	"errors"
)

const WebhookCallbackURL = "https://radiant-mesa-23210.herokuapp.com/bitbucket/"
const BitbucketRepoBase = "https://api.bitbucket.org/2.0/repositories/%v"
//const BitbucketRepoBaseV1 = "https://api.bitbucket.org/1.0/repositories/%s"

//Bitbucket is a bitbucket handler responsible for finding build files and
//registering webhooks for necessary repositories
type Bitbucket struct {
	Client    ocenet.HttpClient
	Marshaler jsonpb.Marshaler

	credConfig	*models.AdminConfig
	isInitialized	bool
}

// Takes in admin config creds, returns any errors that may happen during setup
func (bb *Bitbucket) SetMeUp(adminConfig *models.AdminConfig) error {
	var conf = clientcredentials.Config {
		ClientID:     adminConfig.ClientId,
		ClientSecret: adminConfig.ClientSecret,
		TokenURL:     adminConfig.TokenURL,
	}
	var ctx = context.Background()
	token, err := conf.Token(ctx)
	if err != nil {
		ocelog.IncludeErrField(err).Error("well shit we can't get a token")
		return errors.New("Unable to retrieve token for " + adminConfig.Type + "/" + adminConfig.AcctName)
	}
	ocelog.Log().Debug("token: " + token.AccessToken)

	bitbucketClient := ocenet.HttpClient{}
	bbClient := conf.Client(ctx)
	bitbucketClient.AuthClient = bbClient
	bitbucketClient.Unmarshaler = &jsonpb.Unmarshaler{
		AllowUnknownFields: true,
	}

	//populate fields
	bb.Client = bitbucketClient
	bb.Marshaler = jsonpb.Marshaler{}
	bb.credConfig = adminConfig
	bb.isInitialized = true
	return nil
}

//Walk iterates over all repositories and creates webhook if one doesn't
//exist. Will only work if client has been setup
func (bb Bitbucket) Walk() error {
	if !bb.isInitialized {
		return errors.New("client has not yet been initialized, please call SetMeUp() before walking")
	}
	return bb.recurseOverRepos(fmt.Sprintf(BitbucketRepoBase, bb.credConfig.AcctName))
}

// Get File in repo at a certain commit.
// filepath: string filepath relative to root of repo
// fullRepoName: string account_name/repo_name as it is returned in the Bitbucket api Repo Source `full_name`
// commitHash: string git hash for revision number
func (bb Bitbucket) GetFile(filePath string, fullRepoName string, commitHash string) (bytez []byte, err error) {
	ocelog.Log().Debug("inside GetFile")
	path := fmt.Sprintf("%s/src/%s/%s", fullRepoName, commitHash, filePath)
	bytez, err = bb.Client.GetUrlRawData(fmt.Sprintf(BitbucketRepoBase, path))
	if err != nil {
		return
	}
	return
}

//CreateWebhook will create webhook at specified webhook url
func (bb Bitbucket) CreateWebhook(webhookURL string) error {
	for _, key := range bb.FindWebhooks(webhookURL) {
		//create webhook if one does not already exist
		newWebhook := &pb.CreateWebhook{
			Description: "marianne did this",
			Active:      true,
			Url: WebhookCallbackURL + "/" + key,
			Events: []string{models.BitbucketEvents[key]},
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


//TODO: comment
func (bb Bitbucket) notSetUP() bool {
	return bb.Client == (ocenet.HttpClient{}) || bb.Marshaler == (jsonpb.Marshaler{})
}

//recursively iterates over all repositories and creates webhook
func (bb Bitbucket) recurseOverRepos(repoUrl string) error {
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

//recursively iterates over all webhooks and returns true (matches our callback url) if one already exists
//returns list of event keys that still needs to be created
func (bb Bitbucket) FindWebhooks(getWebhookURL string) []string {
	ocelog.Log().Debug(getWebhookURL)
	needsCreation := []string{}
	if getWebhookURL == "" {
		return needsCreation
	}
	webhooks := &pb.GetWebhooks{}
	bb.Client.GetUrl(getWebhookURL, webhooks)

	if len(webhooks.GetValues()) > 0 {
		for _, wh := range webhooks.GetValues() {
			for k := range models.BitbucketEvents {
				ocelog.Log().Debug(k)
				if wh.GetUrl() != WebhookCallbackURL + "/" + k {
					needsCreation = append(needsCreation, k)
				}
			}
		}
	}




	return bb.FindWebhooks(webhooks.GetNext())
}

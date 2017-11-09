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

const BitbucketRepoBase = "https://api.bitbucket.org/2.0/repositories/%v"
const BitbucketRepoBaseV1 = "https://api.bitbucket.org/1.0/repositories/%s"

//Bitbucket is a bitbucket handler responsible for finding build files and
//registering webhooks for necessary repositories
type Bitbucket struct {
	Client    ocenet.HttpClient
	Marshaler jsonpb.Marshaler

	credConfig	*models.AdminConfig
	isInitialized	bool
}

//TODO: why is up in all lowercase?
// Takes in admin config creds, return true if successful, false if fails
func (bb *Bitbucket) SetMeUP(adminConfig *models.AdminConfig) bool {
	var conf = clientcredentials.Config {
		ClientID:     adminConfig.ClientId,
		ClientSecret: adminConfig.ClientSecret,
		TokenURL:     adminConfig.TokenURL,
	}
	var ctx = context.Background()
	_, err := conf.Token(ctx)
	if err != nil {
		ocelog.LogErrField(err).Error("well shit we can't get a token")
		return false
	}

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
	return true
}

//Walk iterates over all repositories and creates webhook if one doesn't
//exist. Will only work if client has been setup
func (bb Bitbucket) Walk() error {
	if !bb.isInitialized {
		return errors.New("client has not yet been initialized, please call SetMeUP() before walking")
	}
	bb.recurseOverRepos(fmt.Sprintf(BitbucketRepoBase, bb.credConfig.AcctName))
	return nil
}

// Get File in repo at a certain commit.
// filepath: string filepath relative to root of repo
// fullRepoName: string account_name/repo_name as it is returned in the Bitbucket api Repo Source `full_name`
// commitHash: string git hash for revision number
func (bb Bitbucket) GetFile(filePath string, fullRepoName string, commitHash string) (yamlString string, err error) {
	ocelog.Log().Debug("inside GetFile")
	//if bb.notSetUP() == true {
	//	err_str := "cannot get file without cred initialization"
	//	err = errors.New(err_str)
	//	ocelog.Log().Debug(err_str)
	//	return
	//}
	fileResp := &pb.RepoSourceFile{}
	path := fmt.Sprintf("%s/src/%s/%s", fullRepoName, commitHash, filePath)
	bb.Client.GetUrl(fmt.Sprintf(BitbucketRepoBaseV1, path), fileResp)
	yamlString = fileResp.GetData()
	return
}

func (bb Bitbucket) notSetUP() bool {
	return bb.Client == (ocenet.HttpClient{}) || bb.Marshaler == (jsonpb.Marshaler{})
}

//recursively iterates over all repositories and creates webhook
func (bb Bitbucket) recurseOverRepos(repoUrl string) {
	if repoUrl == "" {
		return
	}
	repositories := &pb.PaginatedRepository{}
	bb.Client.GetUrl(repoUrl, repositories)
	for _, v := range repositories.GetValues() {
		fmt.Printf("found repo %v\n", v.GetFullName())
		if !bb.doesWebhookExist(v.GetLinks().GetHooks().GetHref()) {
			//create webhook if one does not already exist
			newWebhook := &pb.CreateWebhook{
				Description: "marianne did this",
				Url:         models.WebhookCallbackURL,
				Active:      true,
				Events:      []string{"repo:push"},
			}
			webhookStr, err := bb.Marshaler.MarshalToString(newWebhook)
			if err != nil {
				ocelog.LogErrField(err).Fatal("failed to convert webhook to json string")
			}
			bb.Client.PostUrl(v.GetLinks().GetHooks().GetHref(), webhookStr, nil)
			ocelog.Log().Debug("subscribed to webhook for ", v.GetLinks().GetHooks().GetHref())
		}
	}
	bb.recurseOverRepos(repositories.GetNext())
}

//recursively iterates over all webhooks and returns true (matches our callback url) if one already exists
func (bb Bitbucket) doesWebhookExist(getWebhookURL string) bool {
	if getWebhookURL == "" {
		return false
	}
	webhooks := &pb.GetWebhooks{}
	bb.Client.GetUrl(getWebhookURL, webhooks)
	for _, wh := range webhooks.GetValues() {
		if wh.GetUrl() == models.WebhookCallbackURL {
			return true
		}
	}
	return bb.doesWebhookExist(webhooks.GetNext())
}

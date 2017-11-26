package handler

//TODO: add interface once we have more than just bitbucket
import (
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/shankj3/ocelot/admin/models"
	"github.com/shankj3/ocelot/util/ocelog"
	"github.com/shankj3/ocelot/util/ocenet"
	pb "github.com/shankj3/ocelot/protos/out"
	"errors"
	"strings"
)

const DefaultCallbackURL = "https://radiant-mesa-23210.herokuapp.com/bitbucket"
const BitbucketRepoBase = "https://api.bitbucket.org/2.0/repositories/%v"

//Bitbucket is a bitbucket handler responsible for finding build files and
//registering webhooks for necessary repositories
type Bitbucket struct {
	CallbackURL	string
	Client    ocenet.HttpClient
	Marshaler jsonpb.Marshaler

	credConfig	*models.AdminConfig
	isInitialized	bool
}

// Takes in admin config creds, returns any errors that may happen during setup
func (bb *Bitbucket) SetMeUp(adminConfig *models.AdminConfig, client ocenet.HttpClient) error {
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
	for _, key := range bb.FindWebhooks(webhookURL) {
		//create webhook if one does not already exist
		newWebhook := &pb.CreateWebhook{
			Description: "marianne did this",
			Active:      true,
			Url: bb.GetCallbackURL() + "/" + key,
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

//GetCallbackURL is a getter for retrieving callbackURL for bitbucket webhooks
func (bb *Bitbucket) GetCallbackURL () string {
	if len(bb.CallbackURL) > 0 {
		return bb.CallbackURL
	}
	return DefaultCallbackURL
}

//SetCallbackURL sets callback urls to be used for webhooks
func (bb *Bitbucket) SetCallbackURL (callbackURL string) {
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
//returns list of event keys that still needs to be created
func (bb *Bitbucket) FindWebhooks(getWebhookURL string) []string {
	var needsCreation []string
	if getWebhookURL == "" {
		return needsCreation
	}
	webhooks := &pb.GetWebhooks{}
	bb.Client.GetUrl(getWebhookURL, webhooks)

	if len(webhooks.GetValues()) > 0 {
		bbEvents := bbEvents(bb.GetCallbackURL())

		for _, wh := range webhooks.GetValues() {
			_, ok := bbEvents[wh.GetUrl()]
			if ok {
				bbEvents[wh.GetUrl()] = true
			}
		}

		for url, evt := range bbEvents {
			if !evt {
				needsCreation = append(needsCreation, strings.TrimPrefix(url, bb.GetCallbackURL() + "/"))
			}
		}
	} else {
		for k := range models.BitbucketEvents {
			ocelog.Log().Debug(k)
			needsCreation = append(needsCreation, k)
		}
	}
	return append(needsCreation, bb.FindWebhooks(webhooks.GetNext())...)
}

//creates a copy of the map of bitbucket events
func bbEvents(callbackURL string) map[string]bool {
	var bbEvents = make(map[string]bool)
	for k, _ := range models.BitbucketEvents {
		bbEvents[callbackURL + "/" + k] = false
	}
	return bbEvents
}

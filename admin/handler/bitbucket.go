package handler

//TODO: add interface once we have more than just bitbucket
import (
	"context"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/shankj3/ocelot/admin/models"
	"github.com/shankj3/ocelot/ocelog"
	"github.com/shankj3/ocelot/ocenet"
	pb "github.com/shankj3/ocelot/protos/out"
	"golang.org/x/oauth2/clientcredentials"
)

const BitbucketRepoBase = "https://api.bitbucket.org/2.0/repositories/%v"

//Bitbucket is a bitbucket handler responsible for finding build files and
//registering webhooks for necessary repositories
type Bitbucket struct {
	Client    ocenet.HttpClient
	Marshaler jsonpb.Marshaler
}

//Subscribe takes in a set of configurations and will kick off
//iterating over repositories
func (bb Bitbucket) Subscribe(adminConfig models.AdminConfig) {
	ocelog.Log.Debug("inside of subscribe", adminConfig)
	var conf = clientcredentials.Config{
		ClientID:     adminConfig.ClientId,
		ClientSecret: adminConfig.ClientSecret,
		TokenURL:     adminConfig.TokenURL,
	}

	var ctx = context.Background()
	token, err := conf.Token(ctx)
	ocelog.Log.Debug("access token: ", token)
	if err != nil {
		ocelog.LogErrField(err).Fatal("well shit we can't get a token")
	}

	bitbucketClient := ocenet.HttpClient{}
	bbClient := conf.Client(ctx)
	bitbucketClient.AuthClient = bbClient
	bitbucketClient.Unmarshaler = &jsonpb.Unmarshaler{
		AllowUnknownFields: true,
	}

	bb.Client = bitbucketClient
	bb.Marshaler = jsonpb.Marshaler{}

	bb.recurseOverRepos(fmt.Sprintf(BitbucketRepoBase, adminConfig.AcctName))
}

//// helper functions for walking repositories and source files ////

func (bb Bitbucket) recurseOverRepos(repoUrl string) {
	if repoUrl == "" {
		return
	}
	repositories := &pb.PaginatedRepository{}
	bb.Client.GetUrl(repoUrl, repositories)
	for _, v := range repositories.GetValues() {
		fmt.Printf("found repo %v\n", v.GetFullName())
		bb.recurseOverFiles(v.GetLinks().GetSource().GetHref(), v.GetLinks().GetHooks().GetHref())
	}
	bb.recurseOverRepos(repositories.GetNext())
}

func (bb Bitbucket) recurseOverFiles(sourceFileUrl string, webhookUrl string) {
	if sourceFileUrl == "" {
		return
	}
	repositories := &pb.PaginatedRootDirs{}
	bb.Client.GetUrl(sourceFileUrl, repositories)
	for _, v := range repositories.GetValues() {
		if v.GetType() == "commit_file" && len(v.GetAttributes()) == 0 && v.GetPath() == models.BuildFileName {
			//found file, subscribe to webhook
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
			bb.Client.PostUrl(webhookUrl, webhookStr, nil)
			ocelog.Log.Debug("subscribed to webhook for ", webhookUrl)

		}
	}
	bb.recurseOverFiles(repositories.GetNext(), webhookUrl)
}

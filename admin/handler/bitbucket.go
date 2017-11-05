package handler

//TODO: could be interface once we have more than just bitbucket
import (
	"github.com/shankj3/ocelot/admin/models"
	"golang.org/x/oauth2/clientcredentials"
	"github.com/shankj3/ocelot/ocelog"
	"github.com/shankj3/ocelot/ocenet"
	"github.com/golang/protobuf/jsonpb"
	pb "github.com/shankj3/ocelot/protos"
	"context"
	"fmt"
)

type Bitbucket struct {
	Client 	ocenet.HttpClient
	Marshaler	jsonpb.Marshaler
}

//TODO: if this is camelcase, how will it be clear this is constant?
const BITBUCKET_REPO_BASE string = "https://api.bitbucket.org/2.0/repositories/%v"


func (bb Bitbucket) Subscribe(adminConfig models.AdminConfig) {

	var conf = clientcredentials.Config{
		ClientID:     adminConfig.ConfigId,
		ClientSecret: adminConfig.ClientSecret,
		TokenURL:     adminConfig.TokenURL,
	}

	//TODO: what the fuck is context
	var ctx = context.Background()

	token, err := conf.Token(ctx)
	ocelog.Log.Debug("access token: ", token.AccessToken)
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

	bb.recurseOverRepos(fmt.Sprintf(BITBUCKET_REPO_BASE, adminConfig.AcctName))
}

func (bb Bitbucket)	recurseOverRepos(repoUrl string) {
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
		if v.GetType() == "commit_file" && len(v.GetAttributes()) == 0 && v.GetPath() == models.BUILD_FILE_NAME {
			//found file, subscribe to webhook
			fmt.Printf("Contains %v\n", v.GetPath())

			newWebhook := &pb.CreateWebhook{
				Description: "marianne did this",
				Url:         models.WEBHOOK_CALLBACK_URL,
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
package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	pb "github.com/shankj3/ocelot/protos/out"
	"golang.org/x/oauth2/clientcredentials"
	"log"
	"net/http"
)

//TODO: read from config file taking in list of bitbucket repos instead
const myClientId string = "wSfJHcfpHP4dzHv8ak"
const myClientSecret string = "jQqXJmn52wupDzqSheckKu62NqpvnFuR"
const l11BitbucketRepo string = "https://api.bitbucket.org/2.0/repositories/level11consulting"

const myTokenURL string = "https://bitbucket.org/site/oauth2/access_token"
const webhookCallbackURL string = ""

//TODO: what the fuck is context
var ctx = context.Background()

//TODO: move all this shit out into its own class that takes creds and provides a client for you? Or maybe just an httputil class that you can use for get/post, etc.
var conf = clientcredentials.Config{
	ClientID:     myClientId,
	ClientSecret: myClientSecret,
	TokenURL:     myTokenURL,
}

var hu = HttpUtil{}

func main() {
	fmt.Println("marianne is number 1\n")
	_, err := conf.Token(ctx)
	if err != nil {
		log.Fatal(err)
	}

	bbClient := conf.Client(ctx)
	hu.BbClient = bbClient
	hu.Unmarshaler = &jsonpb.Unmarshaler{
		AllowUnknownFields: true,
	}

	recurseOverRepos(l11BitbucketRepo)
}

func recurseOverRepos(repoUrl string) {
	if repoUrl == "" {
		return
	}
	repositories := &pb.PaginatedRepository{}
	hu.getRepoPage(repoUrl, repositories)
	for _, v := range repositories.GetValues() {
		fmt.Printf("repo is called %v\n", v.GetFullName())
		recurseOverFiles(v.GetLinks().GetSource().GetHref())
	}
	recurseOverRepos(repositories.GetNext())
}

func recurseOverFiles(sourceFileUrl string) {
	if sourceFileUrl == "" {
		return
	}
	repositories := &pb.PaginatedRootDirs{}
	hu.getRepoPage(sourceFileUrl, repositories)
	for _, v := range repositories.GetValues() {
		if v.GetType() == "commit_file" && len(v.GetAttributes()) == 0 {
			fmt.Printf("found a FILE in the root dir called %v\n", v.GetPath())
		}
	}
	recurseOverFiles(repositories.GetNext())
}

//TODO: probably move to its own util - takes in url with json response and converts to protobuf
type HttpUtil struct {
	BbClient    *http.Client
	Unmarshaler *jsonpb.Unmarshaler
}

func (hu HttpUtil) getRepoPage(repoUrl string, unmarshalObj proto.Message) {
	resp, err := hu.BbClient.Get(repoUrl)
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(resp.Body)

	if err := hu.Unmarshaler.Unmarshal(reader, unmarshalObj); err != nil {
		log.Fatal("failed to parse response from bitbucket", err)
	}
	defer resp.Body.Close()
}

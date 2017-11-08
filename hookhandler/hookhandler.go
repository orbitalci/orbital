package main

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/meatballhat/negroni-logrus"
	"github.com/shankj3/ocelot/admin/handler"
	"github.com/shankj3/ocelot/admin/models"
	"github.com/shankj3/ocelot/nsqpb"
	"github.com/shankj3/ocelot/ocelog"
	pb "github.com/shankj3/ocelot/protos/out"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	"io"
	"net/http"
	"os"
)

// On receive of repo push, marshal the json to an object then write the important fields to protobuf Message on NSQ queue.
func RepoPush(w http.ResponseWriter, r *http.Request) {
	repopush := &pb.RepoPush{}
	HandleUnmarshal(r.Body, repopush)
	buildConf, err := GetBuildConfig(repopush.Repository.FullName, repopush.Push.Changes[0].New.Target.Hash)
	if err != nil {
		// todo: return error message
		ocelog.LogErrField(err).Error("unable to get build conf")
	}
	bundle := &pb.PushBuildBundle{
		Config:     buildConf,
		PushData:   repopush,
		VaultToken: "", // todo: this.
	}
	// send to queue
	if err := nsqpb.WriteToNsq(bundle, nsqpb.PUSH_QUEUE); err != nil {
		ocelog.LogErrField(err).Error("nsq insert webhook error")
	} else {
		ocelog.Log().Info("added repo:push event to nsq", nsqpb.PUSH_QUEUE)
	}
}

func PullRequest(w http.ResponseWriter, r *http.Request) {
	pr := &pb.PullRequest{}
	HandleUnmarshal(r.Body, pr)
	buildConf, err := GetBuildConfig(pr.Pullrequest.Source.Repository.FullName, pr.Pullrequest.Source.Repository.FullName)
	if err != nil {
		// todo: return error message
		ocelog.LogErrField(err).Error("unable to get build conf")
	}

	bundle := &pb.PRBuildBundle{
		Config:     buildConf,
		PrData:     pr,
		VaultToken: "",
	}
	if err := nsqpb.WriteToNsq(bundle, nsqpb.PULL_REQUEST_QUEUE); err != nil {
		ocelog.LogErrField(err).Error("nsq insert webhook error")
	} else {
		ocelog.Log().Info("added pullrequest:created event to nsq", nsqpb.PULL_REQUEST_QUEUE)
	}

}

func HandleUnmarshal(requestBody io.ReadCloser, unmarshalObj proto.Message) {
	unmarshaler := &jsonpb.Unmarshaler{
		AllowUnknownFields: true,
	}
	if err := unmarshaler.Unmarshal(requestBody, unmarshalObj); err != nil {
		ocelog.LogErrField(err).Fatal("could not parse request body into proto.Message")
	}
	defer requestBody.Close()
}

// for testing
// irl... use vault
func getCredConfig() models.AdminConfig {
	return models.AdminConfig{
		ConfigId:     "jessishank",
		ClientId:     "QEBYwP5cKAC3ykhau4",
		ClientSecret: "gKY2S3NGnFzJKBtUTGjQKc4UNvQqa2Vb",
		TokenURL:     "https://bitbucket.org/site/oauth2/access_token",
		AcctName:     "jessishank",
	}
}

func GetBuildConfig(repoFullName string, checkoutCommit string) (conf *pb.BuildConfig, err error) {
	cfg := getCredConfig()
	bb := handler.Bitbucket{}
	bb.SetMeUP(&cfg)
	confstr, err := bb.GetFile("ocelot.yml", repoFullName, checkoutCommit)
	if err != nil {
		return
	}
	conf = &pb.BuildConfig{}
	if err = ConvertYAMLtoProtobuf([]byte(confstr), conf); err != nil {
		return
	}
	return
}

func main() {
	ocelog.InitializeOcelog(ocelog.GetFlags())
	ocelog.Log().Debug()
	port := os.Getenv("PORT")
	if port == "" {
		ocelog.Log().Fatal("$PORT must be set")
	}
	mux := mux.NewRouter()
	mux.HandleFunc("/test", RepoPush).Methods("POST")
	// mux.HandleFunc("/", ViewWebhooks).Methods("GET")

	n := negroni.New(negroni.NewRecovery(), negroni.NewStatic(http.Dir("public")))
	n.Use(negronilogrus.NewCustomMiddleware(ocelog.GetLogLevel(), &log.JSONFormatter{}, "web"))
	n.UseHandler(mux)
	n.Run(":" + port)
}

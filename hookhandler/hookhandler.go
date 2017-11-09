package main

import (
	"encoding/json"
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

const BuildTopic = "repo_build"

// On receive of repo push, marshal the json to an object then write the important fields to protobuf Message on NSQ queue.
func RepoPush(w http.ResponseWriter, r *http.Request) {
	repopush := &pb.RepoPush{}
	if err := HandleUnmarshal(r.Body, repopush); err != nil {
		SetHttpError(w, "could not parse request body into proto.Message", err)
	}

	buildConf, err := GetBuildConfig(repopush.Repository.FullName, repopush.Push.Changes[0].New.Target.Hash)
	if err != nil {
		//ocelog.LogErrField(err).Error("unable to get build conf")
		SetHttpError(w, "unable to get build conf", err)
		return
	}
	// instead, add to topic. each worker gets a topic off a channel,
	// so one worker to one channel
	bundle := &pb.PushBuildBundle{
		Config:     buildConf,
		PushData:   repopush,
		VaultToken: "", // todo: this.
	}

	// send to queue
	if err := nsqpb.WriteToNsq(bundle, BuildTopic); err != nil {
		//ocelog.LogErrField(err).Error("nsq insert webhook error")
		SetHttpError(w, "nsq insert webhook error", err)
		return

	} else {
		ocelog.Log().Info("added repo:push event to nsq", BuildTopic)
	}
}

func PullRequest(w http.ResponseWriter, r *http.Request) {
	pr := &pb.PullRequest{}
	if err := HandleUnmarshal(r.Body, pr); err != nil {
		SetHttpError(w, "could not parse request body into proto.Message", err)
		return
	}
	buildConf, err := GetBuildConfig(pr.Pullrequest.Source.Repository.FullName, pr.Pullrequest.Source.Repository.FullName)
	if err != nil {
		//ocelog.LogErrField(err).Error("unable to get build conf")
		SetHttpError(w, "unable to get build conf", err)
		return
	}

	bundle := &pb.PRBuildBundle{
		Config:     buildConf,
		PrData:     pr,
		VaultToken: "",
	}
	if err := nsqpb.WriteToNsq(bundle, BuildTopic); err != nil {
		//ocelog.LogErrField(err).Error("nsq insert webhook error")
		SetHttpError(w, "nsq insert webhook error", err)
		return
	} else {
		ocelog.Log().Info("added pullrequest:created event to nsq", BuildTopic)
	}

}

func HandleUnmarshal(requestBody io.ReadCloser, unmarshalObj proto.Message) (err error){
	unmarshaler := &jsonpb.Unmarshaler{
		AllowUnknownFields: true,
	}
	if err := unmarshaler.Unmarshal(requestBody, unmarshalObj); err != nil {
		//ocelog.LogErrField(err).Fatal("could not parse request body into proto.Message")
		return
	}
	defer requestBody.Close()
	return
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

type RESTError struct {
	err error
	errorDescription string
}

func SetHttpError(w http.ResponseWriter, error_desc string, err error) {
	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set("Content-Type", "application/json")
	resterr := RESTError{
		err: err,
		errorDescription: error_desc,
	}
	json.NewEncoder(w).Encode(resterr)
}


func main() {
	ocelog.InitializeOcelog(ocelog.GetFlags())
	ocelog.Log().Debug()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
		ocelog.Log().Warning("running on default port :8088")
	}
	muxi := mux.NewRouter()
	muxi.HandleFunc("/test", RepoPush).Methods("POST")
	// mux.HandleFunc("/", ViewWebhooks).Methods("GET")

	n := negroni.New(negroni.NewRecovery(), negroni.NewStatic(http.Dir("public")))
	n.Use(negronilogrus.NewCustomMiddleware(ocelog.GetLogLevel(), &log.JSONFormatter{}, "web"))
	n.UseHandler(muxi)
	n.Run(":" + port)
}

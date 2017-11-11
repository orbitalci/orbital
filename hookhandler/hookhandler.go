package main

import (
	"github.com/gorilla/mux"
	"github.com/meatballhat/negroni-logrus"
	"github.com/shankj3/ocelot/admin/handler"
	"github.com/shankj3/ocelot/admin/models"
	pb "github.com/shankj3/ocelot/protos/out"
	//"github.com/shankj3/ocelot/util/consulet"
	"github.com/shankj3/ocelot/util/deserialize"
	"github.com/shankj3/ocelot/util/nsqpb"
	"github.com/shankj3/ocelot/util/ocelog"
	"github.com/shankj3/ocelot/util/ocenet"
	"github.com/shankj3/ocelot/util/ocevault"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	"net/http"
	"os"
	"sync"
)

const BuildTopic = "repo_build"
//var consul = consulet.Default()
var deserializer = deserialize.New()

var vaultCached *ocevault.Ocevault
var once sync.Once

// the sync.Once() way of letting something get initialized only one time.
func getInitVault() *ocevault.Ocevault {
	once.Do(func() {
		ocev, err := ocevault.NewEnvAuthClient()
		if err != nil {
			ocelog.IncludeErrField(err).Fatal("vault must be initialized.")
		}
		vaultCached = ocev
	})
	return vaultCached
}

// On receive of repo push, marshal the json to an object then write the important fields to protobuf Message on NSQ queue.
func RepoPush(w http.ResponseWriter, r *http.Request) {
	repopush := &pb.RepoPush{}
	if err := deserializer.JSONToProto(r.Body, repopush); err != nil {
		msg := "could not parse request body into proto.Message"
		ocelog.IncludeErrField(err).Error(msg)
		ocenet.JSONApiError(w, http.StatusBadRequest, msg, err)
	}

	buildConf, err := GetBuildConfig(repopush.Repository.FullName, repopush.Push.Changes[0].New.Target.Hash)
	if err != nil {
		msg := "unable to get build conf"
		ocelog.IncludeErrField(err).Error(msg)
		ocenet.JSONApiError(w, http.StatusBadRequest, msg, err)
		return
	}
	vault := getInitVault()
	token, err := vault.CreateThrowawayToken()
	if err != nil {
		msg := "unable to create one-time vault token"
		ocelog.IncludeErrField(err).Error(msg)
		ocenet.JSONApiError(w, http.StatusBadRequest, msg, err)
	}
	// instead, add to topic. each worker gets a topic off a channel,
	// so one worker to one channel
	bundle := &pb.PushBuildBundle{
		Config:     buildConf,
		PushData:   repopush,
		VaultToken: token,
	}
	go nsqpb.WriteToNsq(bundle, BuildTopic)
}

func PullRequest(w http.ResponseWriter, r *http.Request) {
	pr := &pb.PullRequest{}
	if err := deserializer.JSONToProto(r.Body, pr); err != nil {
		ocenet.JSONApiError(w, http.StatusBadRequest, "could not parse request body into proto.Message", err)
		return
	}
	buildConf, err := GetBuildConfig(pr.Pullrequest.Source.Repository.FullName, pr.Pullrequest.Source.Repository.FullName)
	if err != nil {
		//ocelog.IncludeErrField(err).Error("unable to get build conf")
		ocenet.JSONApiError(w, http.StatusBadRequest, "unable to get build conf", err)
		return
	}
	// get one-time token use for access to vault
	vault := getInitVault()
	token, err := vault.CreateThrowawayToken()
	if err != nil {
		msg := "unable to create one-time vault token"
		ocelog.IncludeErrField(err).Error(msg)
		ocenet.JSONApiError(w, http.StatusBadRequest, msg, err)
	}
	// create bundle, send that s*** off!
	bundle := &pb.PRBuildBundle{
		Config:     buildConf,
		PrData:     pr,
		VaultToken: token,
	}
	go nsqpb.WriteToNsq(bundle, BuildTopic)

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
	bb.SetMeUp(&cfg)
	confstr, err := bb.GetFile("ocelot.yml", repoFullName, checkoutCommit)
	if err != nil {
		return
	}
	conf = &pb.BuildConfig{}
	if err = deserializer.YAMLToProto([]byte(confstr), conf); err != nil {
		return
	}
	return
}


func main() {
	ocelog.InitializeOcelog(ocelog.GetFlags())
	ocelog.Log().Debug()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
		ocelog.Log().Warning("running on default port :8088")
	}
	// initialize vault on startup, we want to know right away if we don't have the creds we need.
	_ = getInitVault()
	muxi := mux.NewRouter()
	muxi.HandleFunc("/test", RepoPush).Methods("POST")
	// mux.HandleFunc("/", ViewWebhooks).Methods("GET")

	n := negroni.New(negroni.NewRecovery(), negroni.NewStatic(http.Dir("public")))
	n.Use(negronilogrus.NewCustomMiddleware(ocelog.GetLogLevel(), &log.JSONFormatter{}, "web"))
	n.UseHandler(muxi)
	n.Run(":" + port)
}

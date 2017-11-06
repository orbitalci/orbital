package main

import (
	// "bufio"
	"flag"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/meatballhat/negroni-logrus"
	"github.com/shankj3/ocelot/nsqpb"
	"github.com/shankj3/ocelot/ocelog"
	pb "github.com/shankj3/ocelot/protos"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	"io"
	"net/http"
	"os"
)

// On receive of repo push, marshal the json to an object then write the important fields to protobuf Message on NSQ queue.
func RepoPush(w http.ResponseWriter, r *http.Request) {
	// b, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	//     ocelog.LogErrField(err).Fatal("error reading http request body")
	// }
	repopush := pb.RepoPush{}
	HandleUnmarshal(r.Body, &repopush)
	queue_topic := "repo_push"
	if err := nsqpb.WriteToNsq(&repopush, queue_topic); err != nil {
		ocelog.LogErrField(err).Warn("nsq insert webhook error")
	} else {
		ocelog.Log.Info("added to nsq ", queue_topic)
	}
}

func HandleUnmarshal(requestBody io.ReadCloser, unmarshalObj proto.Message) {
	unmarshaler := &jsonpb.Unmarshaler{
		AllowUnknownFields: true,
	}
	if err := unmarshaler.Unmarshal(requestBody, unmarshalObj); err != nil {
		ocelog.LogErrField(err).Fatal("could not parse repo push")
	}
	defer requestBody.Close()
}

func GetFlags() string {
	// write flag
	var log_level string
	flag.StringVar(&log_level, "log_level", "warn", "set log level")
	flag.Parse()
	return log_level
}

func main() {
	// initialize logger
	// log_level := GetFlags()
	ocelog.InitializeOcelog(GetFlags())
	ocelog.Log.Debug()
	port := os.Getenv("PORT")
	if port == "" {
		ocelog.Log.Fatal("$PORT must be set")
	}
	mux := mux.NewRouter()
	mux.HandleFunc("/test", RepoPush).Methods("POST")
	// mux.HandleFunc("/", ViewWebhooks).Methods("GET")

	n := negroni.New(negroni.NewRecovery(), negroni.NewStatic(http.Dir("public")))
	n.Use(negronilogrus.NewCustomMiddleware(ocelog.GetLogLevel(), &log.JSONFormatter{}, "web"))
	n.UseHandler(mux)
	n.Run(":" + port)
}

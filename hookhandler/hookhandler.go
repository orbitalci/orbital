package main

import (
    "flag"
    "github.com/golang/protobuf/jsonpb"
    "github.com/gorilla/mux"
    "github.com/meatballhat/negroni-logrus"
    "github.com/shankj3/ocelot/ocelog"
    pb "github.com/shankj3/ocelot/protos"
    log "github.com/sirupsen/logrus"
    "github.com/urfave/negroni"
    "net/http"
    "os"
)

// On receive of repo push, marshal the json to an object then write the important fields to protobuf Message on NSQ queue.
func RepoPush(w http.ResponseWriter, r *http.Request) {
    // b, err := ioutil.ReadAll(r.Body)
    // if err != nil {
    //     ocelog.LogErrField(err).Fatal("error reading http request body")
    // }
    unmarshaler := &jsonpb.Unmarshaler{
        AllowUnknownFields: true,
    }
    repopush := &pb.RepoPush{}
    if err := unmarshaler.Unmarshal(r.Body, repopush); err != nil {
        ocelog.LogErrField(err).Fatal("could not unmarshal bitbucket repo push to struct")
    }

    // err = database.AddToPostgres(repopush.Repository.Links.HTML.Href, latestChange.New.Target.Hash)
    if err := WriteToNsq(repopush); err != nil {
        ocelog.LogErrField(err).Warn("nsq insert webhook error")
    }
}

// func ViewWebhooks(w http.ResponseWriter, r *http.Request) {
//     rows := database.PullWebhookFromPostgres()
//     defer rows.Close()
//     for rows.Next() {
//         var repourl string
//         var githash string
//         var hook_time time.Time
//         _ = rows.Scan(&repourl, &githash, &hook_time)
//         w.Write([]byte(database.WriteWebhookString(repourl, githash, hook_time)))
//     }
// }

// Adding in the flags struct now because i'm sure I'll be adding more flags and it would
// become unruly otherwise

type HookHandlerFlags struct {
    log_level string
    // make enum
    //
}

//write flags for this service. Add your flag here
func (self *HookHandlerFlags) writeFlags() {
    flag.StringVar(&self.log_level, "log_level", "warn", "set log level")
}

func (self *HookHandlerFlags) parseFlags() {
    self.writeFlags()
    flag.Parse()
}

func main() {
    h := HookHandlerFlags{}
    h.parseFlags()
    // initialize logger
    ocelog.InitializeOcelog(h.log_level)
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

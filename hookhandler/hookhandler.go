package main

import (
    "encoding/json"
    "flag"
    "io/ioutil"
    "net/http"
    "os"
    "time"

    "github.com/gorilla/mux"
    "github.com/meatballhat/negroni-logrus"
    // "github.com/shankj3/ocelot/hookhandler/database"
    "github.com/shankj3/ocelot/ocelog"
    log "github.com/sirupsen/logrus"
    "github.com/urfave/negroni"
    "gopkg.in/go-playground/webhooks.v3/bitbucket"
    // for pretty printing objects:
    // "github.com/davecgh/go-spew/spew"
)

// On receive of repo push, marshal the json to an object then write the important fields to protobuf Message on NSQ queue.
func RepoPush(w http.ResponseWriter, r *http.Request) {
    b, err := ioutil.ReadAll(r.Body)
    if err != nil {
        ocelog.LogErrField(err).Fatal("error reading http request body")
    }
    repopush := bitbucket.RepoPushPayload{}
    err = json.Unmarshal(b, &repopush)
    if err != nil {
        ocelog.LogErrField(err).Fatal("could not unmarshal bitbucket repo push to struct")
    }
    protoMsg := ConvertHookToProto(repopush)
    // err = database.AddToPostgres(repopush.Repository.Links.HTML.Href, latestChange.New.Target.Hash)
    err = WriteToNsq(&protoMsg)
    if err != nil {
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
    //
    logLevel, _ := log.ParseLevel(h.log_level)
    // initialize logger
    ocelog.InitializeOcelog(logLevel)
    ocelog.Log.Debug()
    port := os.Getenv("PORT")
    if port == "" {
        ocelog.Log.Fatal("$PORT must be set")
    }
    mux := mux.NewRouter()
    mux.HandleFunc("/test", RepoPush).Methods("POST")
    mux.HandleFunc("/", ViewWebhooks).Methods("GET")

    n := negroni.New(negroni.NewRecovery(), negroni.NewStatic(http.Dir("public")))
    n.Use(negronilogrus.NewCustomMiddleware(logLevel, &log.JSONFormatter{}, "web"))
    n.UseHandler(mux)
    n.Run(":" + port)
}

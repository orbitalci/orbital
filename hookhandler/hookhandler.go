package main

import (
    "os"

    "encoding/json"
    "flag"
    "io/ioutil"
    "net/http"
    "time"

    "github.com/gorilla/mux"
    "github.com/meatballhat/negroni-logrus"
    "github.com/shankj3/ocelot/hookhandler/database"
    "github.com/shankj3/ocelot/ocelog"
    log "github.com/sirupsen/logrus"
    "github.com/urfave/negroni"
    "gopkg.in/go-playground/webhooks.v3/bitbucket"
    // for pretty printing objects:
    // "github.com/davecgh/go-spew/spew"
)

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
        ocelog.LogErrField(err).Fatal("nsq insert webhook error")
    }
}

func ViewWebhooks(w http.ResponseWriter, r *http.Request) {
    rows := database.PullWebhookFromPostgres()
    defer rows.Close()
    for rows.Next() {
        var repourl string
        var githash string
        var hook_time time.Time
        _ = rows.Scan(&repourl, &githash, &hook_time)
        w.Write([]byte(database.WriteWebhookString(repourl, githash, hook_time)))
    }
}

func main() {
    var log_level string
    flag.StringVar(&log_level, "log_level", "warn", "set log level")
    flag.Parse()
    //
    logLevel, _ := log.ParseLevel(log_level)
    // initialize logger
    ocelog.InitializeOcelog(logLevel)
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

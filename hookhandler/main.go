package main

import (
    "os"
    "log"
    "time"
    "net/http"
    "io/ioutil"
    "encoding/json"

    "github.com/urfave/negroni"
    "github.com/gorilla/mux"
    "gopkg.in/go-playground/webhooks.v3/bitbucket"

    "github.com/shankj3/hookhandler/database"
    // for pretty printing objects:
    // "github.com/davecgh/go-spew/spew"
)

func RepoPush(w http.ResponseWriter, r *http.Request) {
    b, err := ioutil.ReadAll(r.Body)
    if err != nil {
        log.Println("ERROR:")
    }
    repopush := bitbucket.RepoPushPayload{}
    err1 := json.Unmarshal(b, &repopush)
    if err1 != nil {
        log.Println("Error", err1)
    }
    pgErr := database.AddToPostgres(repopush.Repository.Links.HTML.Href, repopush.Push.Changes[0].New.Target.Hash)
    if pgErr != nil {
         log.Fatal(pgErr)
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
    port := os.Getenv("PORT")
    if port == "" {
        log.Fatal("$PORT must be set")
    }
    mux := mux.NewRouter()
    mux.HandleFunc("/test", RepoPush).Methods("POST")
    mux.HandleFunc("/", ViewWebhooks).Methods("GET")

    n := negroni.Classic()
    n.UseHandler(mux)
    n.Run(":" + port)
}
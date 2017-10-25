package main

import (
    "fmt"
    "net/http"
    "encoding/json"
    "github.com/urfave/negroni"
    "github.com/gorilla/mux"
    "gopkg.in/go-playground/webhooks.v3/bitbucket"
    // for pretty printing objects:
    // "github.com/davecgh/go-spew/spew"
    "io/ioutil"
    // "json"
)

func MyHandler(w http.ResponseWriter, r *http.Request) {
    // vars := mux.Vars(r)
    // fmt.Printf("Vars: %+v \n", vars)
    // fmt.Printf("%+v\n\n", r)
    // spew.Dump(r)
    // fmt.Printf("\n\n\n\n\n")
    b, err := ioutil.ReadAll(r.Body)
    if err != nil {
        fmt.Println("ERROR:")
    }
    repopush := bitbucket.RepoPushPayload{}
    err1 := json.Unmarshal(b, &repopush)
    if err1 != nil {
        fmt.Println("Error", err1)
    }
    fmt.Printf("%s\n", repopush.Push.Changes[0].New.Target.Hash)
    w.Write([]byte("hi"))
}

func main() {
    mux := mux.NewRouter()
    mux.HandleFunc("/test", MyHandler).Methods("POST")
    n := negroni.Classic()
    n.UseHandler(mux)
    n.Run(":3000")
}
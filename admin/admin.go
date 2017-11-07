package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/shankj3/ocelot/admin/handler"
	"github.com/shankj3/ocelot/admin/models"
	"github.com/shankj3/ocelot/ocelog"
	"net/http"
	"os"
)

//TODO: write the part that parses config file

//TODO: this will eventually get moved to secrets and/or consul and not be in memory map
var creds = map[string]models.AdminConfig{}
var configChannel = make(chan models.AdminConfig)

func main() {
	ocelog.InitializeOcelog("debug")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		ocelog.Log().Warn("Running on default port 8080")
	}

	go ListenForConfig()

	mux := mux.NewRouter()
	mux.HandleFunc("/", ConfigHandler).Methods("POST")
	mux.HandleFunc("/", ListConfigHandler).Methods("GET")
	ocelog.Log().Fatal(http.ListenAndServe(":" + port, mux))
}

func ListConfigHandler(w http.ResponseWriter, r *http.Request) {

}

func ConfigHandler(w http.ResponseWriter, r *http.Request) {
	var adminConfig models.AdminConfig
	_ = json.NewDecoder(r.Body).Decode(&adminConfig)
	//TODO: validate config here
	creds[adminConfig.ConfigId] = adminConfig

	//publish config creds to listener
	go func() {
		configChannel <- adminConfig
	}()
}

func ListenForConfig() {
	for config := range configChannel {
		ocelog.Log().Debug("received new config", config)
		go handler.Bitbucket{}.Subscribe(config)
	}
	//TODO: close channel when finished
}
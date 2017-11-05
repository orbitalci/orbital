package main

import (
	"github.com/shankj3/ocelot/ocelog"
	"github.com/shankj3/ocelot/admin/models"
	"github.com/shankj3/ocelot/admin/handler"
	"github.com/gorilla/mux"
	"os"
	"net/http"
	"encoding/json"
)


//TODO: this will eventually get moved to secrets and/or consul and not be in memory map
var creds = map[string]models.AdminConfig{}
var configChannel = make(chan models.AdminConfig)

func main() {
	ocelog.InitializeOcelog("debug")

	port := os.Getenv("PORT")
	if port == "" {
		ocelog.Log.Fatal("$PORT must be set")
	}

	mux := mux.NewRouter()
	mux.HandleFunc("/", ConfigHandler).Methods("POST")
	ocelog.Log.Fatal(http.ListenAndServe(":" + port, mux))

	//TODO: CREATE CLIENTS FOR CONFIG AND RUN IN SEPARATE THREAD - look into thread safety
	for {
		go handler.Bitbucket{}.Subscribe(<- configChannel)
	}

	//TODO: close channel when finished

}

func ConfigHandler(w http.ResponseWriter, r *http.Request) {
	var adminConfig	models.AdminConfig
	_ = json.NewDecoder(r.Body).Decode(&adminConfig)
	//TODO: validate config here
	creds[adminConfig.ConfigId] = adminConfig

	//publish config creds to listener
	configChannel <- adminConfig
}




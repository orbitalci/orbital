package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"gopkg.in/go-playground/validator.v9"
	"github.com/shankj3/ocelot/admin/handler"
	"github.com/shankj3/ocelot/admin/models"
	"github.com/google/uuid"
	"github.com/shankj3/ocelot/ocelog"
	"net/http"
	"os"
	"github.com/shankj3/ocelot/ocenet"
)

//TODO: write the part that parses config file

//TODO: this will eventually get moved to secrets and/or consul and not be in memory map
var creds = map[string]models.AdminConfig{}
var configChannel = make(chan models.AdminConfig)
var validate = validator.New()

func main() {
	ocelog.InitializeOcelog("debug")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		ocelog.Log.Warn("Running on default port 8080")
	}

	go ListenForConfig()

	mux := mux.NewRouter()
	mux.HandleFunc("/", ConfigHandler).Methods("POST")
	mux.HandleFunc("/", ListConfigHandler).Methods("GET")
	ocelog.Log.Fatal(http.ListenAndServe(":" + port, mux))
}

func ListConfigHandler(w http.ResponseWriter, r *http.Request) {

}

func ConfigHandler(w http.ResponseWriter, r *http.Request) {
	var adminConfig models.AdminConfig
	_ = json.NewDecoder(r.Body).Decode(&adminConfig)

	errorMsg, err := validateConfig(&adminConfig)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errorMsg)
		return
	}

	//set the config id if it doesn't exist
	if len(adminConfig.ConfigId) == 0 {
		adminConfig.ConfigId = uuid.New().String()
	}

	creds[adminConfig.ConfigId] = adminConfig
	go func() {
		configChannel <- adminConfig
	}()
}

func ListenForConfig() {
	for config := range configChannel {
		ocelog.Log.Debug("received new config", config)
		go handler.Bitbucket{}.Subscribe(config)
	}
}

func validateConfig(adminConfig *models.AdminConfig) ([]byte, error) {
	err := validate.Struct(adminConfig)
	if err != nil {
		var errorMsg string
		for _, nestedErr := range err.(validator.ValidationErrors) {
			errorMsg = nestedErr.Field() + " is " + nestedErr.Tag()
			ocelog.Log.Warn(errorMsg)
		}

		errJson := &ocenet.HttpError{
			Status: http.StatusBadRequest,
			Error: errorMsg,
		}

		convertedError, nestedErr := json.Marshal(errJson)
		if nestedErr != nil {
			ocelog.LogErrField(err)
		}
		return convertedError, err
	}
	return nil, nil
}
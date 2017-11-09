package main

import (
	"encoding/json"
	"gopkg.in/yaml.v2"
	"github.com/gorilla/mux"
	"gopkg.in/go-playground/validator.v9"
	"github.com/shankj3/ocelot/admin/handler"
	"github.com/shankj3/ocelot/admin/models"
	"github.com/shankj3/ocelot/util/ocenet"
	"github.com/shankj3/ocelot/util/ocelog"
	"net/http"
	"os"
	"io/ioutil"
	"github.com/google/uuid"
)

//TODO: write the part that talks to consul
//TODO: hook admin code into vault
//TODO: look into hookhandler logic and separate into new ocelot.yaml + new commit

//TODO: this will eventually get moved to secrets and/or consul and not be in memory map
var creds = map[string]models.AdminConfig{}
var configChannel = make(chan models.AdminConfig)
var validate = validator.New()

func main() {
	ocelog.InitializeOcelog("debug")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		ocelog.Log().Warn("Running on default port 8080")
	}

	go ListenForConfig()
	ReadConfig()


	mux := mux.NewRouter()
	//TODO: seems like maybe this should be command line tool instead - wait for Abby
		//list all configs
		//list all repos + 'tracked' repos vs. 'untracked' repos
		//add new repo
		//configure whether or not you want admin to discover new ocelot.yaml files for you

	mux.HandleFunc("/", ConfigHandler).Methods("POST")
	mux.HandleFunc("/", ListConfigHandler).Methods("GET")
	ocelog.Log().Fatal(http.ListenAndServe(":" + port, mux))
}

//TODO: change this to stop returning passwords (BLOCKED till vault + consul is done)
func ListConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(creds)
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
	configChannel <- adminConfig
}

//TODO: possibly take this out into a util class where it can convert data from any file into struct? What about protobuf?
//reads config file in current directory if it exists, exits if file is unparseable or doesn't exist
func ReadConfig() {
	config := &models.ConfigYaml{}
	configFile, err := ioutil.ReadFile(models.ConfigFileName)
	if err != nil {
		ocelog.LogErrField(err)
		return
	}
	err = yaml.Unmarshal(configFile, config)
	if err != nil {
		ocelog.LogErrField(err)
		return
	}
	for configKey, configVal := range config.Credentials {
		configVal.ConfigId = configKey
		configChannel <- configVal
	}
}

//when new configurations are added to the config channel, create bitbucket client and webhooks
func ListenForConfig() {
	for config := range configChannel {
		ocelog.Log().Debug("received new config", config)
		handler := handler.Bitbucket{}
		ok := handler.SetMeUP(&config)

		if !ok {
			ocelog.Log().Error("could not setup bitbucket client")
			continue
		}

		go handler.Walk()
		creds[config.ConfigId] = config
	}
}

//validates config and returns json formatted error
func validateConfig(adminConfig *models.AdminConfig) ([]byte, error) {
	err := validate.Struct(adminConfig)
	if err != nil {
		var errorMsg string
		for _, nestedErr := range err.(validator.ValidationErrors) {
			errorMsg = nestedErr.Field() + " is " + nestedErr.Tag()
			ocelog.Log().Warn(errorMsg)
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
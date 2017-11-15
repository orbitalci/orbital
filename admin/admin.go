package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"gopkg.in/go-playground/validator.v9"
	"github.com/shankj3/ocelot/admin/handler"
	"github.com/shankj3/ocelot/admin/models"
	"github.com/shankj3/ocelot/util/ocenet"
	"github.com/shankj3/ocelot/util/ocelog"
	"github.com/shankj3/ocelot/util/consulet"
	"github.com/shankj3/ocelot/util/deserialize"
	"github.com/namsral/flag"
	"net/http"
	"io/ioutil"
	"github.com/google/uuid"
)

//TODO: write the part that talks to consul
//TODO: hook admin code into vault
//TODO: look into hookhandler logic and separate into new ocelot.yaml + new commit


var validate = validator.New()
var consul = consulet.Default()
var deserializer = deserialize.New()

func main() {
	//load properties
	var port string
	var consulHost string
	var consulPort int
	var logLevel string

	flag.StringVar(&port, "port", "8080", "admin server port")
	flag.StringVar(&consulHost, "consul-host", "localhost", "consul host")
	flag.IntVar(&consulPort, "consul-port", 8500, "consul port")
	flag.StringVar(&logLevel, "log-level", "debug", "ocelot admin log level")
	flag.Parse()

	ocelog.InitializeOcelog(logLevel)

	//register to consul
	err := consul.RegisterService("localhost", 8080, "ocelot-admin")
	if err != nil {
		ocelog.LogErrField(err)
	}

	//check for config on load
	ReadConfig()

	//start http server
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
	//creds := map[string]*models.AdminConfig

	creds := map[string]models.AdminConfig{}
	for _, v := range consul.GetKeyValues("creds") {
		appName = v.Key
	}

	json.NewEncoder(w).Encode(creds)
}

//ConfigHandler handles config from REST api
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

	errorMsg, err = SetupCredentials(&adminConfig)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(errorMsg)
		return
	}

}

//reads config file in current directory if it exists, exits if file is unparseable or doesn't exist
func ReadConfig() {
	config := &models.ConfigYaml{}
	configFile, err := ioutil.ReadFile(models.ConfigFileName)
	if err != nil {
		ocelog.LogErrField(err)
		return
	}
	err = deserializer.YAMLToStruct(configFile, config)
	if err != nil {
		ocelog.LogErrField(err)
		return
	}
	for configKey, configVal := range config.Credentials {
		configVal.ConfigId = configKey

		_, err = SetupCredentials(&configVal)
		ocelog.LogErrField(err)
	}
}

//when new configurations are added to the config channel, create bitbucket client and webhooks
func SetupCredentials(config *models.AdminConfig) ([]byte, error) {
	handler := handler.Bitbucket{}
	err := handler.SetMeUp(config)

	if err != nil {
		ocelog.Log().Error("could not setup bitbucket client")

		errJson := &ocenet.HttpError{
			Status: http.StatusUnprocessableEntity,
			Error: "Could not setup bitbucket client for " + config.ConfigId,
			ErrorDetail: err.Error(),
		}

		convertedError, nestedErr := json.Marshal(errJson)
		if nestedErr != nil {
			ocelog.LogErrField(err)
		}
		return convertedError, err
	}

	err = handler.Walk()
	if err != nil {

		errJson := &ocenet.HttpError{
			Status: http.StatusUnprocessableEntity,
			Error: "Could not traverse repositories and create necessary webhooks for " + config.ConfigId,
			ErrorDetail: err.Error(),
		}

		convertedError, nestedErr := json.Marshal(errJson)
		if nestedErr != nil {
			ocelog.LogErrField(err)
		}
		return convertedError, err
	}

	storeConfig(config)
	return nil, nil
}

func storeConfig(config *models.AdminConfig) {
	consul.AddKeyValue("creds/" + config.ConfigId + "/clientid", []byte(config.ClientId))
	//TODO: move the secret into vault
	consul.AddKeyValue("creds/" + config.ConfigId + "/clientsecret", []byte(config.ClientSecret))
	consul.AddKeyValue("creds/" + config.ConfigId + "/tokenurl", []byte(config.TokenURL))
	consul.AddKeyValue("creds/" + config.ConfigId + "/acctname", []byte(config.AcctName))

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
			ErrorDetail: err.Error(),
		}

		convertedError, nestedErr := json.Marshal(errJson)
		if nestedErr != nil {
			ocelog.LogErrField(err)
		}
		return convertedError, err
	}
	return nil, nil
}
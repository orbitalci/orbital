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
	"github.com/shankj3/ocelot/util/ocevault"
	"strings"
)

//TODO: look into hookhandler logic and separate into new ocelot.yaml + new commit
//TODO: rewrite admin code to use grpc
//TODO: floe integration??? just putting this note here so we remember
//TODO: add type to the config object and also post data onto vault using the path + acctname

var validate = validator.New()
var consul = consulet.Default()
var deserializer = deserialize.New()
//TODO: should probably not swallow error and do real initialization
var vault, _ = ocevault.NewEnvAuthClient()

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

	//TODO: do we really give a shit about registering to consul
	err := consul.RegisterService("localhost", 8080, "ocelot-admin")
	if err != nil {
		ocelog.IncludeErrField(err).Error()
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

func ListConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	creds := map[string]*models.AdminConfig{}

	for _, v := range consul.GetKeyValues("creds") {
		consulCreds := strings.Split(strings.TrimLeft(v.Key,"creds/"), "/")
		config, ok := creds[consulCreds[0]]
		if !ok {
			config = &models.AdminConfig{
				ConfigId: consulCreds[0],
				ClientSecret: "**********",
			}
			creds[consulCreds[0]] = config
		}
		switch consulCreds[1] {
			case "clientid":
				config.ClientId = string(v.Value[:])
			case "tokenurl":
				config.TokenURL = string(v.Value[:])
			case "acctname":
				config.AcctName = string(v.Value[:])
			default:
				ocelog.Log().Info("unrecognized consul key")
		}
	}

	json.NewEncoder(w).Encode(creds)
}

//ConfigHandler handles config from REST api
func ConfigHandler(w http.ResponseWriter, r *http.Request) {
	var adminConfig models.AdminConfig
	_ = json.NewDecoder(r.Body).Decode(&adminConfig)

	errorMsg, err := validateConfig(&adminConfig)

	if err != nil {
		ocenet.JSONApiError(w, http.StatusBadRequest, errorMsg, err)
		return
	}

	//set the config id if it doesn't exist
	if len(adminConfig.ConfigId) == 0 {
		adminConfig.ConfigId = uuid.New().String()
	}

	errorMsg, err = SetupCredentials(&adminConfig)
	if err != nil {
		ocenet.JSONApiError(w, http.StatusUnprocessableEntity, errorMsg, err)
		return
	}

}

//reads config file in current directory if it exists, exits if file is unparseable or doesn't exist
func ReadConfig() {
	config := &models.ConfigYaml{}
	configFile, err := ioutil.ReadFile(models.ConfigFileName)
	if err != nil {
		ocelog.IncludeErrField(err).Error()
		return
	}
	err = deserializer.YAMLToStruct(configFile, config)
	if err != nil {
		ocelog.IncludeErrField(err).Error()
		return
	}
	for configKey, configVal := range config.Credentials {
		configVal.ConfigId = configKey

		_, err = SetupCredentials(&configVal)
		ocelog.IncludeErrField(err).Error()
	}
}

//when new configurations are added to the config channel, create bitbucket client and webhooks
func SetupCredentials(config *models.AdminConfig) (string, error) {
	handler := handler.Bitbucket{}
	err := handler.SetMeUp(config)

	if err != nil {
		ocelog.Log().Error("could not setup bitbucket client")
		return "Could not setup bitbucket client for " + config.ConfigId, err
	}

	err = handler.Walk()
	if err != nil {
		return "Could not traverse repositories and create necessary webhooks for " + config.ConfigId, err
	}

	storeConfig(config)
	return "", nil
}

func storeConfig(config *models.AdminConfig) {
	consul.AddKeyValue("creds/" + config.ConfigId + "/clientid", []byte(config.ClientId))
	//TODO: move the secret into vault
	consul.AddKeyValue("creds/" + config.ConfigId + "/ClientSecret", []byte(config.ClientSecret))
	consul.AddKeyValue("creds/" + config.ConfigId + "/tokenurl", []byte(config.TokenURL))
	consul.AddKeyValue("creds/" + config.ConfigId + "/acctname", []byte(config.AcctName))

}

//validates config and returns json formatted error
func validateConfig(adminConfig *models.AdminConfig) (string, error) {
	err := validate.Struct(adminConfig)
	if err != nil {
		var errorMsg string
		for _, nestedErr := range err.(validator.ValidationErrors) {
			errorMsg = nestedErr.Field() + " is " + nestedErr.Tag()
			ocelog.Log().Warn(errorMsg)
		}
		return errorMsg, err
	}
	return "", nil
}
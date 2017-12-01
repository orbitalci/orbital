package main

import (
	"encoding/json"
	"github.com/namsral/flag"
	"github.com/shankj3/ocelot/admin/handler"
	"github.com/shankj3/ocelot/admin/models"
	"github.com/shankj3/ocelot/util/deserialize"
	"github.com/shankj3/ocelot/util/ocelog"
	"github.com/shankj3/ocelot/util/ocenet"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"io/ioutil"
	"net/http"
	"github.com/shankj3/ocelot/util"
	"gopkg.in/go-playground/validator.v9"
	"fmt"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"net"
	"google.golang.org/grpc/credentials"
	"crypto/tls"
	"strings"
	"log"
	"crypto/x509"
	"github.com/philips/grpc-gateway-example/insecure"
)

//TODO: rewrite hookhandler logic to use remoteconfig
//TODO: rewrite admin code to use grpc

//TODO: floe integration??? just putting this note here so we remember
//TODO: rewrite all the objects so that they use CredWrapper + Credentials protobuf classes instead - SHIIIIIIT HALFWAY THROUGH
//TODO: change this to use my fork of logrus so we can pretty print logs


//application context, contains stuff that'll get used across admin code
type AdminCtx struct {
	deserializer	*deserialize.Deserializer
	adminValidator	*AdminValidator
	remoteConfig	*util.RemoteConfig
}

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

	//TODO: this is my local vault root token, too lazy to set env variable
	configInstance, err := util.GetInstance(consulHost, consulPort, "0fd3c0f4-52ec-7d3b-d29b-4b4df1326ded")

	if err != nil {
		ocelog.Log().Fatal("could not talk to consul or vault, bailing")
	}

	//populate admin context
	adminContext := &AdminCtx{
		deserializer: deserialize.New(),
		adminValidator: GetValidator(),
		remoteConfig: configInstance,
	}

	//check for config on load
	ReadConfig(adminContext)

	fakeCert := x509.NewCertPool()
	ok := fakeCert.AppendCertsFromPEM([]byte(insecure.Cert))
	//ok := fakeCert.AppendCertsFromPEM([]byte(models.Cert))
	if !ok {
		panic("bad certs")
	}

	pair, err := tls.X509KeyPair([]byte(insecure.Cert), []byte(insecure.Key))
	//pair, err := tls.X509KeyPair([]byte(models.Cert), []byte(models.Key))
	fakeKeyPair := &pair

	//grpc server
	opts := []grpc.ServerOption{
		grpc.Creds(credentials.NewClientTLSFromCert(fakeCert, "localhost:10000"))}

	grpcServer := grpc.NewServer(opts...)
	models.RegisterGuideOcelotServer(grpcServer, NewGuideOcelotServer())
	ctx := context.Background()

	dcreds := credentials.NewTLS(&tls.Config{
		ServerName: "localhost:10000",
		RootCAs:    fakeCert,
	})
	dopts := []grpc.DialOption{grpc.WithTransportCredentials(dcreds)}

	mux := http.NewServeMux()

	//grpc gateway proxy
	gwmux := runtime.NewServeMux()
	err = models.RegisterGuideOcelotHandlerFromEndpoint(ctx, gwmux, "localhost:10000", dopts)
	if err != nil {
		fmt.Printf("serve: %v\n", err)
		return
	}

	mux.Handle("/", gwmux)

	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", 10000))
	if err != nil {
		panic(err)
	}

	srv := &http.Server{
		Addr:    "localhost:10000",
		Handler: grpcHandlerFunc(grpcServer, mux),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*fakeKeyPair},
			NextProtos:   []string{"h2"},
		},
	}

	err = srv.Serve(tls.NewListener(conn, srv.TLSConfig))

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	//list all configs
	//list all repos + 'tracked' repos vs. 'untracked' repos
	//add new repo
	//configure whether or not you want admin to discover new ocelot.yaml files for you

	//TODO: change these two lines below to use grpc gateway instead
	//mux.Handle("/", &ocenet.AppContextHandler{adminContext, ConfigHandler}).Methods("POST")
	//mux.Handle("/", &ocenet.AppContextHandler{adminContext,ListConfigHandler}).Methods("GET")
	//ocelog.Log().Fatal(http.ListenAndServe(":"+port, mux))
}

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise. Copied from cockroachdb.
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is a partial recreation of gRPC's internal checks https://github.com/grpc/grpc-go/pull/514/files#diff-95e9a25b738459a2d3030e1e6fa2a718R61
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			ocelog.Log().Debug("serving grpc guuuuurl")
			grpcServer.ServeHTTP(w, r)
		} else {
			ocelog.Log().Debug("serving HTTTTTTTTTTPPPPPP")
			otherHandler.ServeHTTP(w, r)
		}
	})
}

//func ListConfigHandler(ctx interface{}, w http.ResponseWriter, r *http.Request) {
//	adminCtx := ctx.(*AdminCtx)
//
//	w.Header().Set("Content-Type", "application/json")
//	creds, _ := adminCtx.remoteConfig.GetCredAt(util.ConfigPath, true)
//	json.NewEncoder(w).Encode(creds)
//}

//ConfigHandler handles config from REST api
func ConfigHandler(ctx interface{}, w http.ResponseWriter, r *http.Request) {
	adminCtx := ctx.(*AdminCtx)

	var adminConfig models.AdminConfig
	_ = json.NewDecoder(r.Body).Decode(&adminConfig)

	errorMsg, err := adminCtx.adminValidator.ValidateConfig(&adminConfig)

	if err != nil {
		ocenet.JSONApiError(w, http.StatusBadRequest, errorMsg, err)
		return
	}

	errorMsg, err = SetupCredentials(adminCtx, &adminConfig)
	if err != nil {
		ocenet.JSONApiError(w, http.StatusUnprocessableEntity, errorMsg, err)
		return
	}

}

//reads config file in current directory if it exists, exits if file is unparseable or doesn't exist
func ReadConfig(ctx *AdminCtx) {
	config := &models.ConfigYaml{}
	configFile, err := ioutil.ReadFile(models.ConfigFileName)
	if err != nil {
		ocelog.IncludeErrField(err).Error()
		return
	}
	err = ctx.deserializer.YAMLToStruct(configFile, config)
	if err != nil {
		ocelog.IncludeErrField(err).Error()
		return
	}
	for _, configVal := range config.Credentials {
		errMsg, err := ctx.adminValidator.ValidateConfig(&configVal)
		if err != nil {
			ocelog.IncludeErrField(err).Error(errMsg)
			continue
		}

		_, err = SetupCredentials(ctx, &configVal)
		if err != nil {
			ocelog.IncludeErrField(err).Error()
		}
	}
}

//when new configurations are added to the config channel, create bitbucket client and webhooks
func SetupCredentials(ctx *AdminCtx, config *models.AdminConfig) (string, error) {
	//hehe right now we only have bitbucket
	switch config.Type {
	case "bitbucket":
		bbHandler := handler.Bitbucket{}
		bitbucketClient := &ocenet.OAuthClient{}
		bitbucketClient.Setup(config)

		err := bbHandler.SetMeUp(config, bitbucketClient)

		if err != nil {
			ocelog.Log().Error("could not setup bitbucket client")
			return "Could not setup bitbucket client for " + config.Type + "/" + config.AcctName, err
		}

		err = bbHandler.Walk()
		if err != nil {
			return "Could not traverse repositories and create necessary webhooks for " + config.Type + "/" + config.AcctName, err
		}
	}
	//configPath := util.ConfigPath + "/" + config.Type + "/" + config.AcctName
	//err := ctx.remoteConfig.AddCreds(configPath, config)
	return "", nil
}


////everything below this is for validating/////

func GetValidator() *AdminValidator {
	adminValidator := &AdminValidator {
		Validate: validator.New(),
	}
	adminValidator.Validate.RegisterValidation("validtype", typeValidation)
	return adminValidator
}


//validator for all admin related stuff
type AdminValidator struct {
	Validate	*validator.Validate
}

//validates config and returns json formatted error
func(adminValidator AdminValidator) ValidateConfig(adminConfig *models.AdminConfig) (string, error) {
	err := adminValidator.Validate.Struct(adminConfig)
	if err != nil {
		var errorMsg string
		for _, nestedErr := range err.(validator.ValidationErrors) {
			errorMsg = nestedErr.Field() + " is " + nestedErr.Tag()
			if nestedErr.Tag() == "validtype" {
				errorMsg = "type must be one of the following: bitbucket"
			}

			ocelog.Log().Warn(errorMsg)
		}
		return errorMsg, err
	}
	return "", nil
}

func typeValidation(fl validator.FieldLevel) bool {
	switch fl.Field().String() {
	case "bitbucket":
		return true
	}
	return false
}
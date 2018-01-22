package main

import (
	"bitbucket.org/level11consulting/go-til/deserialize"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	hh "bitbucket.org/level11consulting/ocelot/hookhandler"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"github.com/gorilla/mux"
	"os"
	"strconv"
	"strings"
	"bitbucket.org/level11consulting/ocelot/client/validate"
)

func main() {
	//ocelog.InitializeLog("debug")
	ocelog.InitializeLog(ocelog.GetFlags())
	ocelog.Log().Debug()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
		ocelog.Log().Warning("running on default port :8088")
	}

	consulHost := os.Getenv("CONSUL_HOST")
	if consulHost == "" {
		consulHost = "localhost"
		ocelog.Log().Warning("consul is assumed to be running on localhost")
	}
	consulPort := os.Getenv("CONSUL_PORT")
	if consulPort == "" {
		consulPort = "8500"
		ocelog.Log().Warning("consul is assumed to be running on port 8500")
	}

	consulPortInt, _ := strconv.Atoi(consulPort)
	remoteConfig, err := cred.GetInstance(consulHost, consulPortInt, "")
	if err != nil {
		ocelog.Log().Fatal(err)
	}

	var hookHandlerContext hh.HookHandler

	mode := os.Getenv("ENV")
	if strings.EqualFold(mode, "dev") {
		hookHandlerContext = &hh.MockHookHandlerContext{}
		hookHandlerContext.SetRemoteConfig(&hh.MockRemoteConfig{})

	} else {
		hookHandlerContext = &hh.HookHandlerContext{}
		hookHandlerContext.SetRemoteConfig(remoteConfig)
	}

	hookHandlerContext.SetDeserializer(deserialize.New())
	hookHandlerContext.SetProducer(nsqpb.GetInitProducer())
	hookHandlerContext.SetValidator(validate.GetOcelotValidator())

	muxi := mux.NewRouter()

	// handleBBevent can take push/pull/ w/e
	muxi.Handle("/bitbucket", &ocenet.AppContextHandler{hookHandlerContext, hh.HandleBBEvent}).Methods("POST")
	n := ocenet.InitNegroni("hookhandler", muxi)
	n.Run(":" + port)
}
package main

import (
	"bitbucket.org/level11consulting/go-til/deserialize"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	"bitbucket.org/level11consulting/ocelot/client/validate"
	hh "bitbucket.org/level11consulting/ocelot/hookhandler"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"github.com/gorilla/mux"
	"github.com/namsral/flag"
	"os"
	"strings"
)

func main() {
	//ocelog.InitializeLog("debug")
	var consulHost, loglevel string
	var consulPort int
	flrg := flag.NewFlagSet("werker", flag.ExitOnError)
	flrg.StringVar(&consulHost, "consul-host", "localhost", "host / ip that consul is running on")
	flrg.StringVar(&loglevel, "log-level", "info", "log level")
	flrg.IntVar(&consulPort, "consul-port", 8500, "port that consul is running on")
	flrg.Parse(os.Args[1:])

	ocelog.InitializeLog(loglevel)
	ocelog.Log().Debug()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
		ocelog.Log().Warning("running on default port :8088")
	}

	remoteConfig, err := cred.GetInstance(consulHost, consulPort, "")
	if err != nil {
		ocelog.Log().Fatal(err)
	}

	var hookHandlerContext hh.HookHandler

	mode := os.Getenv("ENV")
	if strings.EqualFold(mode, "dev") {
		hookHandlerContext = &hh.MockHookHandlerContext{}
		hookHandlerContext.SetRemoteConfig(&hh.MockRemoteConfig{})
		ocelog.Log().Info("hookhandler running in dev mode")

	} else {
		hookHandlerContext = &hh.HookHandlerContext{}
		hookHandlerContext.SetRemoteConfig(remoteConfig)
	}

	hookHandlerContext.SetDeserializer(deserialize.New())
	hookHandlerContext.SetProducer(nsqpb.GetInitProducer())
	// todo: add check for hookHandlerContext being valid
	hookHandlerContext.SetValidator(validate.GetOcelotValidator())
	muxi := mux.NewRouter()

	// handleBBevent can take push/pull/ w/e
	muxi.Handle("/bitbucket", &ocenet.AppContextHandler{hookHandlerContext, hh.HandleBBEvent}).Methods("POST")
	n := ocenet.InitNegroni("hookhandler", muxi)
	n.Run(":" + port)
}
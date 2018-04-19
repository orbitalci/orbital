package main

import (
	"bitbucket.org/level11consulting/go-til/deserialize"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	signal "bitbucket.org/level11consulting/ocelot/build_signaler"
	cred "bitbucket.org/level11consulting/ocelot/common/credentials"
	hh "bitbucket.org/level11consulting/ocelot/router/hookhandler"
	"bitbucket.org/level11consulting/ocelot/version"
	"github.com/gorilla/mux"
	"github.com/namsral/flag"
	"os"
	"bitbucket.org/level11consulting/ocelot/build"
)


func main() {
	//ocelog.InitializeLog("debug")
	defaultName, _ := os.Hostname()

	var consulHost, loglevel, name string
	var consulPort int
	flrg := flag.NewFlagSet("hookhandler", flag.ExitOnError)

	flrg.StringVar(&name, "name", defaultName, "if wish to identify as other than hostname")
	flrg.StringVar(&consulHost, "consul-host", "localhost", "host / ip that consul is running on")
	flrg.StringVar(&loglevel, "log-level", "info", "log level")
	flrg.IntVar(&consulPort, "consul-port", 8500, "port that consul is running on")
	flrg.Parse(os.Args[1:])
	version.MaybePrintVersion(flrg.Args())
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

	//mode := os.Getenv("ENV")
	//if strings.EqualFold(mode, "dev") {
	//	hookHandlerContext = &hh.MockHookHandlerContext{}
	//	hookHandlerContext.SetRemoteConfig(&hh.MockRemoteConfig{})
	//	ocelog.Log().Info("hookhandler running in dev mode")
	//
	//} else {
	hookHandlerContext = &hh.HookHandlerContext{Signaler: &signal.Signaler{}}
	hookHandlerContext.SetRemoteConfig(remoteConfig)
	//}

	hookHandlerContext.SetDeserializer(deserialize.New())
	hookHandlerContext.SetProducer(nsqpb.GetInitProducer())
	hookHandlerContext.SetValidator(build.GetOcelotValidator())
	store, err := hookHandlerContext.GetRemoteConfig().GetOcelotStorage()
	if err != nil {
		ocelog.IncludeErrField(err).Fatal("couldn't get storage!")
	}
	hookHandlerContext.SetStorage(store)
	defer store.Close()

	startServer(hookHandlerContext, port)
}

func startServer(ctx interface{}, port string) {
	muxi := mux.NewRouter()

	// handleBBevent can take push/pull/ w/e
	muxi.Handle("/bitbucket", &ocenet.AppContextHandler{ctx, hh.HandleBBEvent}).Methods("POST")
	n := ocenet.InitNegroni("hookhandler", muxi)
	n.Run(":" + port)
}
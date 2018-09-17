package main

import (
	"github.com/gorilla/mux"
	"github.com/namsral/flag"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shankj3/go-til/deserialize"
	ocelog "github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"
	"github.com/shankj3/go-til/nsqpb"
	"github.com/shankj3/ocelot/build"
	signal "github.com/shankj3/ocelot/build_signaler"
	cred "github.com/shankj3/ocelot/common/credentials"
	hh "github.com/shankj3/ocelot/router/hookhandler"
	"github.com/shankj3/ocelot/version"
	"os"
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

	//mode := os.Getenv("ENV")
	//if strings.EqualFold(mode, "dev") {
	//	hookHandlerContext = &hh.MockHookHandlerContext{}
	//	hookHandlerContext.SetRemoteConfig(&hh.MockRemoteConfig{})
	//	ocelog.Log().Info("hookhandler running in dev mode")
	//
	//} else {
	store, err := remoteConfig.GetOcelotStorage()
	if err != nil {
		ocelog.IncludeErrField(err).Fatal("couldn't get storage!")
	}
	signaler := &signal.Signaler{RC: remoteConfig, Deserializer: deserialize.New(), Producer: nsqpb.GetInitProducer(), OcyValidator: build.GetOcelotValidator(), Store: store}
	teller := &signal.CCWerkerTeller{}
	hookHandlerContext := hh.GetContext(signaler, teller)
	defer store.Close()

	startServer(hookHandlerContext, port)
}

func startServer(ctx *hh.HookHandlerContext, port string) {
	muxi := mux.NewRouter()

	// handleBBevent can take push/pull/ w/e
	muxi.HandleFunc("/bitbucket", ctx.HandleBBEvent).Methods("POST")
	muxi.Handle("/metrics", promhttp.Handler())
	n := ocenet.InitNegroni("hookhandler", muxi)
	n.Run(":" + port)
}

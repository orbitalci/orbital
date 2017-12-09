package main

import (
	"os"
	"github.com/shankj3/ocelot/util/cred"
	"bitbucket.org/level11consulting/go-til/deserialize"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	"github.com/gorilla/mux"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	hh "github.com/shankj3/ocelot/hookhandler"
)

func main() {
	ocelog.InitializeLog(ocelog.GetFlags())
	ocelog.Log().Debug()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
		ocelog.Log().Warning("running on default port :8088")
	}

	remoteConfig, err := cred.GetInstance("", 0, "")
	if err != nil {
		ocelog.Log().Fatal(err)
	}

	hookHandlerContext := &hh.HookHandlerContext{
		RemoteConfig: remoteConfig,
		Deserializer: deserialize.New(),
		Producer:     nsqpb.GetInitProducer(),
	}

	muxi := mux.NewRouter()

	// handleBBevent can take push/pull/ w/e
	muxi.Handle("/bitbucket", &ocenet.AppContextHandler{hookHandlerContext, hh.HandleBBEvent}).Methods("POST")

	// mux.HandleFunc("/", ViewWebhooks).Methods("GET")
	n := ocenet.InitNegroni("hookhandler", muxi)
	n.Run(":" + port)
}
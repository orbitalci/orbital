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
	"net/http"
	"io/ioutil"
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
	muxi.HandleFunc("/marianne", PrintBody).Methods("POST")

	// mux.HandleFunc("/", ViewWebhooks).Methods("GET")
	n := ocenet.InitNegroni("hookhandler", muxi)
	n.Run(":" + port)
}

//this is just temp cause I gotta test some stuff
func PrintBody(w http.ResponseWriter, r *http.Request) {
	respBytes, _ := ioutil.ReadAll(r.Body)
	ocelog.Log().Debug(string(respBytes))
}

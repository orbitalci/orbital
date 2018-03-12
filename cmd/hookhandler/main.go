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
	"time"
)


func listen(p *nsqpb.ProtoConsume, consulHost string, consulPort int, topic, hhName string) {
	for {
		if !nsqpb.LookupTopic(p.Config.LookupDAddress(), topic) {
			time.Sleep(100 * time.Millisecond)
			//time.Sleep(10 * time.Second)
		} else {
			consumeConfig, err := cred.GetInstance(consulHost, consulPort, "")
			if err != nil {
				ocelog.Log().Fatal(err)
			}

			handler := &hh.BuildHookHandler {
				RemoteConfig: consumeConfig,
				Deserializer: deserialize.New(),
				Validator: validate.GetOcelotValidator(),
				Producer: nsqpb.GetInitProducer(),
			}
			p.Handler = handler
			p.ConsumeMessages(topic, hhName)
			ocelog.Log().Info("Consuming messages for topic ", topic)
			break
		}
	}
}

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
	hookHandlerContext.SetValidator(validate.GetOcelotValidator())

	go startServer(hookHandlerContext, port)
	//subscribe to build topic for builds triggered via command line
	var consumers []*nsqpb.ProtoConsume
	supportedTopics := []string{"build_please"}
	for _, topic := range supportedTopics {
		protoConsume := nsqpb.NewDefaultProtoConsume()
		go listen(protoConsume, consulHost, consulPort, topic, name)
		consumers = append(consumers, protoConsume)
	}

	for _, consumer := range consumers {
		<-consumer.StopChan
	}
}

func startServer(ctx interface{}, port string) {
	muxi := mux.NewRouter()

	// handleBBevent can take push/pull/ w/e
	muxi.Handle("/bitbucket", &ocenet.AppContextHandler{ctx, hh.HandleBBEvent}).Methods("POST")
	n := ocenet.InitNegroni("hookhandler", muxi)
	n.Run(":" + port)
}
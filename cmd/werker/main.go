/*
Worker needs to:

Pull off of NSQ Queue
Process config file
run build in docker container
provide results endpoint, way for server to access data
  - do this by implementing what's in github.com/gorilla/websocket/examples/command, using websockets
------

## socket / result streaming
- when build starts w/ id by git_hash, it has channels for stdout & stderr
- werker will have service that lists builds it is running
- on build, new path will be added (http://<werker>:9090/<git_hash> that serves stream over websocket
- admin page with build info will have javascript that reads off socket, writes to view.

## docker build vs kubernetes build

*/

package main

import (
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	"bitbucket.org/level11consulting/ocelot/util/buildruntime"
	"bitbucket.org/level11consulting/ocelot/util/nsqwatch"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"bitbucket.org/level11consulting/ocelot/werker"
	"bitbucket.org/level11consulting/ocelot/werker/builder"
	"bitbucket.org/level11consulting/ocelot/werker/valet"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"bitbucket.org/level11consulting/ocelot/werker/config"
)

//listen will listen for messages for a specified topic. If a message is received, a
//message worker handler is created to process the message
func listen(p *nsqpb.ProtoConsume, topic string, conf *config.WerkerConf, streamingChan chan *werker.Transport, buildChan chan *werker.BuildContext, bv *valet.Valet, store storage.OcelotStorage) {
	for {
		if !nsqpb.LookupTopic(p.Config.LookupDAddress(), topic) {
			ocelog.Log().Debug("i am about to sleep for 10s because i couldn't find the topic at ", p.Config.LookupDAddress())
			time.Sleep(10 * time.Second)
		} else {
			mode := os.Getenv("ENV")
			ocelog.Log().Debug("I AM ABOUT TO LISTEN part 2")
			basher := &builder.Basher{LoopbackIp:conf.LoopBackIp}
			if strings.EqualFold(mode, "dev") { //in dev mode, we download zip from werker
				basher.SetBbDownloadURL(conf.LoopBackIp + ":9090/dev")
			}

			handler := werker.NewWorkerMsgHandler(topic, conf, basher, store, bv, streamingChan, buildChan)
			p.Handler = handler
			p.ConsumeMessages(topic, "werker")
			ocelog.Log().Info("Consuming messages for topic ", topic)
			break
		}
	}
}


func main() {
	conf, err := config.GetConf()
	if err != nil {
		fmt.Errorf("cannot get configuration, exiting.... error: %s", err)
		return
	}
	ocelog.InitializeLog(conf.LogLevel)
	streamingTunnel := make(chan *werker.Transport)
	buildCtxTunnel := make(chan *werker.BuildContext)

	ocelog.Log().Debug("starting up worker on off channels w/ ", conf.WerkerName)


	store, err := conf.RemoteConfig.GetOcelotStorage()
	if err != nil {
		ocelog.IncludeErrField(err).Fatal("COULD NOT GET OCELOT STORAGE! BAILING!")
	}
	consulet := conf.RemoteConfig.GetConsul()
	uuid, err := buildruntime.Register(consulet, conf.RegisterIP, conf.GrpcPort, conf.ServicePort)
	if err != nil {
		ocelog.IncludeErrField(err).Fatal("unable to register werker with consul, this is vital. BAILING!")
	}
	conf.WerkerUuid = uuid
	// kick off ctl-c signal handling
	buildValet := valet.NewValet(conf.RemoteConfig, conf.WerkerUuid, conf.WerkerType)
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		buildValet.SignalRecvDed()
	}()

	// start protoConsumers
	var protoConsumers []*nsqpb.ProtoConsume
	//you should know what channels to subscribe to
	supportedTopics := []string{"build"}

	for _, topic := range supportedTopics {
		protoConsume := nsqpb.NewDefaultProtoConsume()
		// todo: add in ability to change number of concurrent processes handling requests; right now it will just take the nsqpb default of 5
		// eg:
		//   protoConsume.Config.MaxInFlight = GetFromEnv
		ocelog.Log().Debug("I AM ABOUT TO LISTEN")
		go listen(protoConsume, topic, conf, streamingTunnel, buildCtxTunnel, buildValet, store)
		protoConsumers = append(protoConsumers, protoConsume)
	}
	go nsqwatch.WatchAndPause(60, protoConsumers, conf.RemoteConfig, store) // todo: put interval in conf
	go werker.ServeMe(streamingTunnel, buildCtxTunnel, conf, store)
	for _, consumer := range protoConsumers {
		<-consumer.StopChan
	}
}



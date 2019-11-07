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
	ocelog "github.com/shankj3/go-til/log"
	"github.com/shankj3/go-til/nsqpb"

	"sync"

	"github.com/level11consulting/orbitalci/build/builder/shell"
	werkerevent "github.com/level11consulting/orbitalci/build/buildeventhandler/werker"
	"github.com/level11consulting/orbitalci/build/helpers/messageservice"
	"github.com/level11consulting/orbitalci/build/helpers/stringbuilder/workingdir"
	"github.com/level11consulting/orbitalci/build/buildmonitor"
	"github.com/level11consulting/orbitalci/models"
	"github.com/level11consulting/orbitalci/server/grpc/werker"
	"github.com/level11consulting/orbitalci/server/monitor/circuitbreaker"
	"github.com/level11consulting/orbitalci/storage"

	"fmt"
	"os"
	"os/signal"

	//"strings"
	"syscall"
	"time"
)

//listen will listen for messages for a specified topic. If a message is received, a
//message worker handler is created to process the message
func listen(p *nsqpb.ProtoConsume, topic string, conf *WerkerConf, streamingChan chan *models.Transport, buildChan chan *models.BuildContext, bm *buildmonitor.BuildMonitor, store storage.OcelotStorage) {
	for {
		if !nsqpb.LookupTopic(p.Config.LookupDAddress(), topic) {
			ocelog.Log().Info("i am about to sleep for 10s because i couldn't find the topic "+topic+" at ", p.Config.LookupDAddress())
			time.Sleep(10 * time.Second)
		} else {
			//mode := os.Getenv("ENV")
			ocelog.Log().Debug("I AM ABOUT TO LISTEN part 2")


			// Maintainer notes
			// The purpose of this function call is to help future code switching in the "Worker message handler" between
			// * bitbucket and github
			// * and different builder type contexts (docker, exec, etc)
			// We should completely redo this configuration implementation so it isn't spaghetti.

			bshr, err := shell.NewBasher("", "", conf.LoopbackIp, workingdir.GetOcyPrefixFromWerkerType(conf.WerkerType))
			// if couldn't make a new basher, just panic
			if err != nil {
				panic("couldnt' create instance of basher, bailing: " + err.Error())
			}
			//if strings.EqualFold(mode, "dev") { //in dev mode, we download zip from werker
			//	bshr.SetBbDownloadURL(conf.LoopbackIp + ":9090/dev")
			//}

			handler := werkerevent.NewWorkerMsgHandler(topic, conf.WerkerFacts, bshr, store, bm, conf.RemoteConfig, streamingChan, buildChan)
			p.Handler = handler
			p.ConsumeMessages(topic, "werker")
			ocelog.Log().Info("Consuming messages for topic ", topic)
			break
		}
	}
}

func main() {
	conf, err := GetConf()
	if err != nil {
		fmt.Printf("cannot get configuration, exiting.... error: %s\n", err)
		return
	}
	ocelog.InitializeLog(conf.LogLevel)
	streamingTunnel := make(chan *models.Transport)
	buildCtxTunnel := make(chan *models.BuildContext)

	ocelog.Log().Debug("starting up worker on off channels w/ ", conf.WerkerName)

	store, err := conf.RemoteConfig.GetOcelotStorage()
	if err != nil {
		ocelog.IncludeErrField(err).Fatal("COULD NOT GET OCELOT STORAGE! BAILING!")
	}
	consulet := conf.RemoteConfig.GetConsul()
	uuid, err := buildmonitor.Register(consulet, conf.RegisterIP, conf.GrpcPort, conf.ServicePort, conf.tags)

	ocelog.Log().Debug("Werker UUID is ", uuid)
	
	if err != nil {
		ocelog.IncludeErrField(err).Fatal("unable to register werker with consul, this is vital. BAILING!")
	}
	conf.Uuid = uuid
	// kick off ctl-c signal handling
	buildMonitor := buildmonitor.NewBuildMonitor(conf.RemoteConfig, conf.Uuid, conf.WerkerType, store, conf.Ssh)
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		buildMonitor.SignalRecvDed()
	}()
	// start protoConsumers
	var protoConsumers []*nsqpb.ProtoConsume
	supportedTopics := messageservice.GetTopics(conf.tags)
	ocelog.Log().Debug("topics!", supportedTopics)

	for _, topic := range supportedTopics {
		protoConsume := nsqpb.NewDefaultProtoConsume()
		// todo: add in ability to change number of concurrent processes handling requests; right now it will just take the nsqpb default of 5
		// eg:
		//   protoConsume.Config.MaxInFlight = GetFromEnv
		ocelog.Log().Debug("I AM ABOUT TO LISTEN")
		go listen(protoConsume, topic, conf, streamingTunnel, buildCtxTunnel, buildMonitor, store)
		protoConsumers = append(protoConsumers, protoConsume)
	}
	go circuitbreaker.WatchAndPause(int64(conf.HealthCheckInterval), protoConsumers, conf.RemoteConfig, store, conf.DiskUtilityHealthCheck)
	go werker.ServeMe(streamingTunnel, conf.WerkerFacts, store, buildMonitor.BuildReaper)
	go buildMonitor.ListenBuilds(buildCtxTunnel, sync.Mutex{})
	for _, consumer := range protoConsumers {
		<-consumer.StopChan
	}
}

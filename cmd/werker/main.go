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
	"fmt"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"github.com/shankj3/ocelot/werker"
	"time"
)

func retry(p *nsqpb.ProtoConsume, topic string, conf *werker.WerkerConf, tunnel chan *werker.Transport) {
	for {
		if !nsqpb.LookupTopic(p.Config.LookupDAddress(), topic) {
			time.Sleep(10 * time.Second)
		} else {
			handler := &werker.WorkerMsgHandler{
				Topic:    topic,
				WerkConf: conf,
				ChanChan: tunnel,
			}
			p.Handler = handler
			p.ConsumeMessages(topic, conf.WerkerName)
			ocelog.Log().Info("Consuming messages for topic ", topic)
		}
	}
}

func main() {
	conf, err := werker.GetConf()
	if err != nil {
		fmt.Errorf("cannot get configuration, exiting.... error: %s", err)
		return
	}
	ocelog.InitializeLog(conf.LogLevel)
	tunnel := make(chan *werker.Transport)
	ocelog.Log().Debug("starting up worker on off channels w/ ", conf.WerkerName)
	var consumers []*nsqpb.ProtoConsume
    for _, topic := range nsqpb.SupportedTopics {
		protoConsume := nsqpb.NewProtoConsume()
		if nsqpb.LookupTopic(protoConsume.Config.LookupDAddress(), topic) {
			handler := &werker.WorkerMsgHandler{
				Topic:    topic,
				WerkConf: conf,
				ChanChan: tunnel,
			}
			protoConsume.Handler = handler
			protoConsume.ConsumeMessages(topic, conf.WerkerName)
			ocelog.Log().Info("Consuming messages for topic ", topic)
		} else {
			ocelog.Log().Warnf("Topic with name %s not found. Will retry every 10 seconds.", topic)
			//todo: dedupe
			go retry(protoConsume, topic, conf, tunnel)
		}
		consumers = append(consumers, protoConsume)
	}
	go werker.ServeMe(tunnel, conf)
	for _, consumer := range consumers {
		<-consumer.StopChan
	}
}
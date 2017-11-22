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
	"github.com/shankj3/ocelot/util/nsqpb"
	"github.com/shankj3/ocelot/util/ocelog"
	"time"
)

func retry(p *nsqpb.ProtoConsume, topic string, conf *WerkerConf) {
	for {
		if !nsqpb.LookupTopic(p.Config.LookupDAddress(), topic) {
			time.Sleep(10 * time.Second)
		} else {
			handler := &WorkerMsgHandler{
				topic:    topic,
				werkConf: conf,
			}
			p.Handler = handler
			p.ConsumeMessages(topic, conf.werkerName)
			ocelog.Log().Info("Consuming messages for topic ", topic)
		}
	}
}

func main() {
	conf, err := GetConf()
	if err != nil {
		fmt.Errorf("cannot get configuration, exiting.... error: %s", err)
		return
	}
	ocelog.InitializeOcelog(conf.logLevel)
	tunnel := make(chan *Transport)
	ocelog.Log().Debug("starting up worker on off channels w/ ", conf.werkerName)
	var consumers []*nsqpb.ProtoConsume
    for _, topic := range nsqpb.SupportedTopics {
		protoConsume := nsqpb.NewProtoConsume()
		if nsqpb.LookupTopic(protoConsume.Config.LookupDAddress(), topic) {
			handler := &WorkerMsgHandler{
				topic:    topic,
				werkConf: conf,
				chanChan: tunnel,
			}
			protoConsume.Handler = handler
			protoConsume.ConsumeMessages(topic, conf.werkerName)
			ocelog.Log().Info("Consuming messages for topic ", topic)
		} else {
			ocelog.Log().Warnf("Topic with name %s not found. Will retry every 10 seconds.", topic)
			retry(protoConsume, topic, conf)
		}
		consumers = append(consumers, protoConsume)
	}
	go ServeMe(tunnel, conf)
	for _, consumer := range consumers {
		<-consumer.StopChan
	}
}
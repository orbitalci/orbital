/*
Worker needs to:

Pull off of NSQ Queue
Process config file
run build in docker container
provide results endpoint, way for server to access data
*/

package main

import (
    "github.com/shankj3/ocelot/util/nsqpb"
	"github.com/shankj3/ocelot/util/ocelog"
	"os"
	"time"
)

func retry(p *nsqpb.ProtoConsume, topic string, hostname string) {
	for {
		if !nsqpb.LookupTopic(p.Config.LookupDAddress(), topic) {
			time.Sleep(10 * time.Second)
		} else {
			handler := NewWorkerMsgHandler(topic)
			p.Handler = handler
			p.ConsumeMessages(topic, hostname)
			ocelog.Log().Info("Consuming messages for topic ", topic)
		}
	}
}

func main() {
	ocelog.InitializeOcelog(ocelog.GetFlags())

	hostname, err := os.Hostname()
    if err != nil {
    	panic("no hostname for machine!") // may not be right approach...
	}
	ocelog.Log().Debug("starting up worker on hostname ", hostname)
	var consumers []*nsqpb.ProtoConsume
    for _, topic := range nsqpb.SupportedTopics {
		protoConsume := nsqpb.NewProtoConsume()
		if nsqpb.LookupTopic(protoConsume.Config.LookupDAddress(), topic) {
			handler := NewWorkerMsgHandler(topic)
			protoConsume.Handler = handler
			protoConsume.ConsumeMessages(topic, hostname)
			ocelog.Log().Info("Consuming messages for topic ", topic)
		} else {
			ocelog.Log().Warnf("Topic with name %s not found. Will retry every 10 seconds.", topic)
			retry(protoConsume, topic, hostname)
		}
		consumers = append(consumers, protoConsume)
	}
	for _, consumer := range consumers {
		<-consumer.StopChan
	}
}
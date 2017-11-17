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
)


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
			consumers = append(consumers, protoConsume)
		} else {
			// maybe this should just error out completely if the topics we expect aren't there?
			// or retry later?
			// it seems to block
			ocelog.Log().Warnf("Topic with name % not found.", topic)
		}
	}
	for _, consumer := range consumers {
		<-consumer.StopChan
	}
}

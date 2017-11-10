package nsqpb

import (
    "fmt"
    "github.com/golang/protobuf/proto"
    "github.com/nsqio/go-nsq"
    "github.com/shankj3/ocelot/util/ocelog"
    "os"
)

var config = nsq.NewConfig()

// Write Protobuf Message to an NSQ topic with name topicName
// Gets the ip of the NSQ daemon from either the environment variable
//  `NSQD_IP` or sets it to 127.0.0.1. the NSQ daemon should run alongside
// any service that produces messages, so this will work usually.
func WriteToNsq(message proto.Message, topicName string) error {
    // there *should* be a nsqd running at same place as service posting.
    var ip_address string
    if ip_address = os.Getenv("NSQD_IP"); ip_address == "" {
        // todo: this wouldn't work with docker-compose, hsould they just have to set
        // the NSQD_IP?
        ip_address = "127.0.0.1"
    }
    p, err := nsq.NewProducer(fmt.Sprintf("%s:4150", ip_address), config)
    if err != nil {
        ocelog.IncludeErrField(err).Fatal("producer create error")
        return err
    }
    p.SetLogger(NewNSQLoggerAtLevel(ocelog.GetLogLevel()))

    var data []byte
    data, err = proto.Marshal(message)
    if err != nil {
        ocelog.IncludeErrField(err).Warn("proto marshal error")
        return err
    }
    err = p.Publish(topicName, data)
    return err
}

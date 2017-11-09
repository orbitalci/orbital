package nsqpb

import (
    "fmt"
    "github.com/nsqio/go-nsq"
    "github.com/shankj3/ocelot/ocelog"
    "os"
    "sync"
)

// ProtoConsume wraps nsq.Message so that code outside the package can just add a UnmarshalProtoFunc
// that doesn't require messing with nsq fields. just write a function that unmarshals to your proto object
// and does work
// ...put in WORK.
type ProtoConsume struct {
    UnmarshalProtoFunc func([]byte) error
}

// Actual wrapper for UnmarshalProtoFunc --> nsq.HandlerFunc
func (p *ProtoConsume) NSQProtoConsume(msg *nsq.Message) error {
    if err := p.UnmarshalProtoFunc(msg.Body); err != nil {
        ocelog.LogErrField(err).Warn("nsq proto consume error")
        return err
    }
    return nil
}

// Consume messages on a given topic / channel in NSQ protoconsume's UnmarshalProtoFunc will be added with
// a wrapper as a handler for the consumer. The ip address of the NSQLookupd instance
// can be set by the environment variable NSQLOOKUPD_IP, but will default to 127.0.0.1
func (p *ProtoConsume) ConsumeMessages(topicName string, channelName string) error {
    wg := &sync.WaitGroup{}
    wg.Add(1)

    decodeConfig := nsq.NewConfig()
    c, err := nsq.NewConsumer(topicName, channelName, decodeConfig)
    if err != nil {
        ocelog.LogErrField(err).Warn("cannot create nsq consumer")
        return err
    }

    c.SetLogger(NewNSQLoggerAtLevel(ocelog.GetLogLevel()))
    c.AddHandler(nsq.HandlerFunc(p.NSQProtoConsume))

    // NSQLOOKUPD_IP may have to be looked up more than nsqd_ip, since nsqlookupd
    // likely isn't running everywhere.
    var ip_address string
    if ip_address = os.Getenv("NSQLOOKUPD_IP"); ip_address == "" {
        ip_address = "127.0.0.1"
    }

    if err = c.ConnectToNSQLookupd(fmt.Sprintf("%s:4161", ip_address)); err != nil {
        ocelog.LogErrField(err).Warn("cannot connect to nsq")
        return err
    }
    wg.Wait()
    return nil
}
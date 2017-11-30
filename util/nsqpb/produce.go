package nsqpb

import (
	"github.com/golang/protobuf/proto"
    "github.com/nsqio/go-nsq"
    "github.com/shankj3/ocelot/util/ocelog"
	"sync"
)


type PbProduce struct {
	config 	 	*nsq.Config
	Producer 	*nsq.Producer
	nsqpbConfig *NsqConfig
}

func DefaultProducer() (producer *PbProduce, err error){
	producer = &PbProduce{
		config: 	 nsq.NewConfig(),
		nsqpbConfig: DefaultNsqConf(),
	}
	producer.Producer, err = nsq.NewProducer(producer.nsqpbConfig.NsqDAddress(), producer.config)
	if err != nil {
		return
	}
	producer.Producer.SetLogger(NewNSQLoggerAtLevel(ocelog.GetLogLevel()))
	return
}

// Write Protobuf Message to an NSQ topic with name topicName
// Gets the ip of the NSQ daemon from either the environment variable
//  `NSQD_IP` or sets it to 127.0.0.1. the NSQ daemon should run alongside
// any service that produces messages, so this will work usually.
func (p *PbProduce) WriteToNsq(message proto.Message, topicName string) error {
    var data []byte
    data, err := proto.Marshal(message)
    if err != nil {
        ocelog.IncludeErrField(err).Warn("proto marshal error")
        return err
    }
    ocelog.Log().Debug("publishing data to ", topicName)
    err = p.Producer.Publish(topicName, data)
    if err != nil {
    	ocelog.IncludeErrField(err).Error("could not publish to nsq!")
	}
    return err
}


// use this to get a producer instance in your code, it will call only once. need to have global variable
// once and cachedProducer set in your service, then pass those to this.
// look into sync.Once if confused
func GetInitProducer(once sync.Once, cachedProducer *PbProduce) *PbProduce {
	once.Do(func(){
		first, err := DefaultProducer()
		if err != nil {
			ocelog.IncludeErrField(err).Fatal("Producer must be initialized.")
		}
		cachedProducer = first
	})
	return cachedProducer
}
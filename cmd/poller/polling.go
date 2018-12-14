package main

import (
	"fmt"
	"os"
	"time"

	"net/url"

	"github.com/level11consulting/orbitalci/build/buildeventhandler/pull/poll"
	"github.com/level11consulting/orbitalci/server/config"
	"github.com/level11consulting/orbitalci/storage"
	"github.com/level11consulting/orbitalci/version"
	"github.com/namsral/flag"
	ocelog "github.com/shankj3/go-til/log"
	"github.com/shankj3/go-til/nsqpb"
)

// FIXME: consistency: consul's host and port, the var name for configInstance/rc
func configure() config.CVRemoteConfig {
	var loglevel, consuladdr string
	var consulport int
	flrg := flag.NewFlagSet("poller", flag.ExitOnError)
	flrg.StringVar(&loglevel, "log-level", "info", "log level")
	flrg.StringVar(&consuladdr, "consul-host", "localhost", "address of consul")
	flrg.IntVar(&consulport, "consul-port", 8500, "port of consul")
	flrg.Parse(os.Args[1:])
	version.MaybePrintVersion(flrg.Args())
	ocelog.InitializeLog(loglevel)
	ocelog.Log().Debug()

	parsedConsulURL, parsedErr := url.Parse(fmt.Sprintf("consul://%s:%d", consuladdr, consulport))
	if parsedErr != nil {
		ocelog.IncludeErrField(parsedErr).Fatal("failed parsing consul uri, bailing")
	}

	rc, err := config.GetInstance(parsedConsulURL, "")
	// todo: add getaddress() to consuletty
	//ocelog.Log().Debug("consul address is ", rc.GetConsul().Config.Address)
	if err != nil {
		ocelog.IncludeErrField(err).Fatal("unable to get instance of remote config, exiting")
	}
	return rc
}

func loadFromDb(store storage.OcelotStorage) error {
	oldPolls, err := store.GetAllPolls()
	if err != nil {
		return err
	}
	for _, oldPoll := range oldPolls {
		if err = poll.WriteCronFile(oldPoll, poll.CronDir); err != nil {
			ocelog.IncludeErrField(err).Error("couldn't write old cron files")
		}

	}
	return err
}

func main() {
	rc := configure()
	supportedTopics := []string{"poll_please", "no_poll_please"}
	store, err := rc.GetOcelotStorage()
	if err != nil {
		ocelog.IncludeErrField(err).Fatal("unable to get ocelot storage, bailing")
	}
	defer store.Close()
	if err = loadFromDb(store); err != nil {
		ocelog.IncludeErrField(err).Fatal("unable to load old cron files from db, bailing")
	}
	// todo: do we need signal recovery here? wouldn't be bad to just put back on the queue
	var consumers []*nsqpb.ProtoConsume
	for _, topic := range supportedTopics {
		protoConsume := nsqpb.NewDefaultProtoConsume()
		go consume(protoConsume, topic)
		consumers = append(consumers, protoConsume)
	}
	fmt.Println(consumers)
	for _, consumer := range consumers {
		<-consumer.StopChan
	}

}

func consume(p *nsqpb.ProtoConsume, topic string) {
	for {
		if !nsqpb.LookupTopic(p.Config.LookupDAddress(), topic) {
			ocelog.Log().Info("about to sleep for 10s because could not find topic ", topic)
			time.Sleep(10 * time.Second)
		} else {
			ocelog.Log().Info("about to consume messages for topic ", topic)
			handler := poll.NewMsgHandler(topic)
			p.Handler = handler
			p.ConsumeMessages(topic, "poller")
			ocelog.Log().Info("consuming messages for topic ", topic)
			break
		}
	}
}

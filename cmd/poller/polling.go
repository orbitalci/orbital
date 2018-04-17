package main

import (
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	"bitbucket.org/level11consulting/ocelot/build_signaler/poll"
	"bitbucket.org/level11consulting/ocelot/models/pb"
	cred "bitbucket.org/level11consulting/ocelot/common/credentials"
	"bitbucket.org/level11consulting/ocelot/storage"
	"fmt"
	"bitbucket.org/level11consulting/ocelot/version"
	"github.com/namsral/flag"
	"os"
	"time"
)

func configure() cred.CVRemoteConfig {
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
	rc, err := cred.GetInstance(consuladdr, consulport, "")
	ocelog.Log().Debug("consul address is ", rc.GetConsul().Config.Address)
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
		msg := &pb.PollRequest{
			Account: oldPoll.Account,
			Repo: oldPoll.Repo,
			Cron: oldPoll.Cron,
			Branches: oldPoll.Branches,
		}
		if err = poll.WriteCronFile(msg); err != nil {
			ocelog.IncludeErrField(err).Error("couldn't write old cron files")
		}

	}
	return err
}

func main() {
	rc := configure()
	supportedTopics := []string{"poll_please", "no_poll_please"}
	store, err := rc.GetOcelotStorage()
	defer store.Close()
	if err != nil {
		ocelog.IncludeErrField(err).Fatal("unable to get ocelot storage, bailing")
	}
	if err = loadFromDb(store); err != nil {
		ocelog.IncludeErrField(err).Fatal("unable to load old cron files from db, bailing")
	}
	// todo: do we need signal recovery here? wouldn't be bad to just put back on the queue
	var consumers []*nsqpb.ProtoConsume
	for _, topic := range supportedTopics {
		protoConsume := nsqpb.NewDefaultProtoConsume()
		go consume(protoConsume, topic, store)
		consumers = append(consumers, protoConsume)
	}
	fmt.Println(consumers)
	for _, consumer := range consumers {
		<-consumer.StopChan
	}

}

func consume(p *nsqpb.ProtoConsume, topic string, store storage.PollTable) {
	for {
		if !nsqpb.LookupTopic(p.Config.LookupDAddress(), topic) {
			ocelog.Log().Info("about to sleep for 10s because could not find topic ", topic)
			time.Sleep(10 * time.Second)
		} else {
			ocelog.Log().Info("about to consume messages for topic ", topic)
			handler := &poll.MsgHandler{Topic: topic, Store: store}
			p.Handler = handler
			p.ConsumeMessages(topic, "poller")
			ocelog.Log().Info("consuming messages for topic ", topic)
			break
		}
	}
}
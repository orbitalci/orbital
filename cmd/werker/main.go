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
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"bitbucket.org/level11consulting/ocelot/werker"
	"bitbucket.org/level11consulting/ocelot/werker/builder"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"io"
	"os"
	"strings"
	"time"
)

//listen will listen for messages for a specified topic. If a message is received, a
//message worker handler is created to process the message
func listen(p *nsqpb.ProtoConsume, topic string, conf *werker.WerkerConf, tunnel chan *werker.Transport, store storage.OcelotStorage) {
	for {
		if !nsqpb.LookupTopic(p.Config.LookupDAddress(), topic) {
			time.Sleep(10 * time.Second)
		} else {
			mode := os.Getenv("ENV")
			basher := &builder.Basher{}
			if strings.EqualFold(mode, "dev") { //in dev mode, we download zip from werker
				basher.SetBbDownloadURL("docker.for.mac.localhost:9090/dev")
			}

			handler := &werker.WorkerMsgHandler{
				Topic:    topic,
				WerkConf: conf,
				ChanChan: tunnel,
				Basher: basher,
				Store:  store,
			}
			p.Handler = handler
			p.ConsumeMessages(topic, conf.WerkerName)
			ocelog.Log().Info("Consuming messages for topic ", topic)
			break
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
	//you should know what channels to subscribe to
	supportedTopics := []string{"build"}

	//do whatever setup stuff werker needs in this function
	setupWerker()
	store, err := conf.RemoteConfig.GetOcelotStorage()
	if err != nil {
		ocelog.IncludeErrField(err).Fatal("COULD NOT GET OCELOT STORAGE! BAILING!")
	}

	//TODO: worker message handler would parse env, if in dev mode, create dev basher and set
	for _, topic := range supportedTopics {
		protoConsume := nsqpb.NewProtoConsume()
		go listen(protoConsume, topic, conf, tunnel, store)
		consumers = append(consumers, protoConsume)
	}

	go werker.ServeMe(tunnel, conf, store)
	for _, consumer := range consumers {
		<-consumer.StopChan
	}
}

//performs whatever setup is needed by werker, right now copies over bb_download.sh to $HOME/.ocelot
//
//***WARINING**** this assumes you're inside of ocelot root dir
func setupWerker() {
	pwd, _ := os.Getwd()
	downloadFile, err := os.Open(pwd + "/template/bb_download.sh")
	if err != nil {
		ocelog.IncludeErrField(err).Error("failed to open code download file")
		return
	}
	defer downloadFile.Close()

	destFile, _ := homedir.Expand("~/.ocelot/bb_download.sh")

	// just get rid of old file
	err = os.Remove(destFile)
	if err != nil {
		ocelog.IncludeErrField(err).Error("failed to remove old file at ~/.ocelot/bb_download.sh")
	}
	ocelog.Log().Info("removed old bb_download")

	destDownloadFile, err := os.Create(destFile)
	if err != nil {
		ocelog.IncludeErrField(err).Error("failed to create file at ~/.ocelot/bb_download.sh")
		return
	}
	defer destDownloadFile.Close()

	if _, err = io.Copy(destDownloadFile, downloadFile); err != nil {
		ocelog.IncludeErrField(err).Error("failed to copy file to ~/.ocelot/bb_download.sh")
		return
	}

	err = os.Chmod(destFile, 0555)
	if err != nil {
		ocelog.IncludeErrField(err).Error("could not change file to be executable")
		return
	}
	ocelog.Log().Info("successfully installed bb_download.sh for use in containers.")
}
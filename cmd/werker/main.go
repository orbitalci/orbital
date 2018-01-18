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
	"bitbucket.org/level11consulting/ocelot/werker"
	"fmt"
	"time"
	"bitbucket.org/level11consulting/ocelot/werker/builder"
	"os"
	"github.com/mitchellh/go-homedir"
	"io"
	"strings"
)

//listen will listen for messages for a specified topic. If a message is received, a
//message worker handler is created to process the message
func listen(p *nsqpb.ProtoConsume, topic string, conf *werker.WerkerConf, tunnel chan *werker.Transport) {
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

	//TODO: worker message handler would parse env, if in dev mode, create dev basher and set
	for _, topic := range supportedTopics {
		protoConsume := nsqpb.NewProtoConsume()
		go listen(protoConsume, topic, conf, tunnel)
		consumers = append(consumers, protoConsume)
	}

	go werker.ServeMe(tunnel, conf)
	for _, consumer := range consumers {
		<-consumer.StopChan
	}
}

//performs whatever setup is needed by werker, right now copies over bb_download.sh to $HOME/.ocelot
//***WARINING**** this assumes you're inside of cmd/werker folder
func setupWerker() {
	pwd, _ := os.Getwd()

	//TODO: this may eventually iterate over template directory and copy all files to .ocelot AND DON'T HARDCODE FILE NAME!!!
	downloadFile, err := os.Open(pwd + "/template/bb_download.sh")
	if err != nil {
		ocelog.IncludeErrField(err).Error("failed to open code download file")
		return
	}
	defer downloadFile.Close()

	destFile, _ := homedir.Expand("~/.ocelot/bb_download.sh")

	_, err = os.Stat(destFile)
	if !os.IsNotExist(err) {
		err = os.Remove(destFile)
		if err != nil {
			ocelog.IncludeErrField(err).Error("failed to remove file at ~/.ocelot/bb_download.sh")
			return
		}
	}


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
}
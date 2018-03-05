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
	"bitbucket.org/level11consulting/ocelot/util/buildruntime"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"bitbucket.org/level11consulting/ocelot/werker"
	"bitbucket.org/level11consulting/ocelot/werker/builder"
	"bitbucket.org/level11consulting/ocelot/werker/recovery"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"
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
			basher := &builder.Basher{LoopbackIp:conf.LoopBackIp}
			if strings.EqualFold(mode, "dev") { //in dev mode, we download zip from werker
				basher.SetBbDownloadURL(conf.LoopBackIp + ":9090/dev")
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
	consulet := conf.RemoteConfig.GetConsul()
	uuid, err := buildruntime.Register(consulet, conf.RegisterIP, conf.GrpcPort, conf.ServicePort)
	if err != nil {
		ocelog.IncludeErrField(err).Fatal("unable to register werker with consul, this is vital. BAILING!")
	}
	conf.WerkerUuid = uuid
	// kick off ctl-c signal handling
	recov := recovery.NewRecovery(conf.RemoteConfig, conf.WerkerUuid)
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		ocelog.Log().Info("received interrupt, cleaning up after myself...")
		recov.StoreInterrupt(recovery.Signal)
		recov.Cleanup()
		os.Exit(1)
	}()
	// start consumers
	//TODO: worker message handler would parse env, if in dev mode, create dev basher and set
	for _, topic := range supportedTopics {
		protoConsume := nsqpb.NewDefaultProtoConsume()
		// todo: add in ability to change number of concurrent processes handling requests; right now it will just take the nsqpb default of 5
		// eg:
		//   protoConsume.Config.MaxInFlight = GetFromEnv
		go listen(protoConsume, topic, conf, tunnel, store)
		consumers = append(consumers, protoConsume)
	}

	go werker.ServeMe(tunnel, conf, store)
	for _, consumer := range consumers {
		<-consumer.StopChan
	}
}

//performs whatever setup is needed by werker, right now copies over everything in cmd/werker/templates to $HOME/.ocelot
// this no longer requires starting werker from a specific place, runtime.Caller(0) knows where this main.go is, and we
// can build off of that
//TODO: this does not work when werker is spawned inside of a docker container. My fix is to make it not exit, bu just log. Profusely.
func setupWerker() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		ocelog.Log().Error("could not call runtime.Caller? this has never happened before")
	}

	templdir := path.Join(path.Dir(filename), "template")
	files, err := ioutil.ReadDir(templdir)
	if err != nil {
		ocelog.IncludeErrField(err).Error(fmt.Sprintf("unable to read directory at: %s", templdir))
	}


	for _, file := range files {
		if file.IsDir() {continue}
		downloadFP := path.Join(templdir, file.Name())
		destFile, err := homedir.Expand("~/.ocelot/"+file.Name())
		if err != nil {
			ocelog.IncludeErrField(err).Error("unable to expand homedir")
		}
		err = addFileToWerker(downloadFP, destFile)
		if err != nil {
			ocelog.IncludeErrField(err).Error("unable to create file ", destFile)
		}
	}
}

func addFileToWerker(originPath string, destFile string) (err error) {
	downloadFile, err := os.Open(originPath)
	if err != nil {
		ocelog.IncludeErrField(err).Error("failed to open file at ", originPath)
		return
	}
	defer downloadFile.Close()
	// just get rid of old file
	err = os.Remove(destFile)
	if err != nil {
		ocelog.IncludeErrField(err).Error("failed to remove old file at ", destFile)
	}
	ocelog.Log().Info("removed old file at ", destFile)

	destDownloadFile, err := os.Create(destFile)
	if err != nil {
		ocelog.IncludeErrField(err).Error("failed to create file at ", destFile)
		return
	}
	defer destDownloadFile.Close()
	if _, err = io.Copy(destDownloadFile, downloadFile); err != nil {
		ocelog.IncludeErrField(err).Error("failed to copy file to ", destFile)
		return
	}
	err = os.Chmod(destFile, 0555)
	if err != nil {
		ocelog.IncludeErrField(err).Error("could not change file to be executable")
		return
	}
	ocelog.Log().Info("successfully installed ", destFile, " for use in containers.")
	return
}
package werker

import (
	ocelog "bitbucket.org/level11consulting/go-til/log"
	d "bitbucket.org/level11consulting/go-til/deserialize"
	"leveler/server"
	"bytes"
	"encoding/gob"
	"bufio"
	"log"
)

// Transport struct is for the Transport channel that will interact with the streaming side of the service
// to stream results back to the admin. It sends just enough to be unique, the hash that triggered the build
// and the InfoChan which the builder will write to.
type Transport struct {
	Hash       string
	InfoChan   chan []byte
}

type WorkerMsgHandler struct {
	Topic    string
	WerkConf *WerkerConf
	infochan chan []byte
	ChanChan chan *Transport
	Deserializer d.Deserializer
}

type WerkerTask struct {
	VaultToken   string
	CheckoutHash string
	Pipe         *server.PipelineConfig
}

// UnmarshalAndProcess is called by the nsq consumer to handle the build message
func (w WorkerMsgHandler) UnmarshalAndProcess(msg []byte) error {
	ocelog.Log().Debug("unmarshaling build obj and processing")

	// channels get closed after the build finishes
	w.infochan = make(chan []byte)
	// set goroutine for watching for results and logging them (for rn)
	// cant add go watchForResults here bc can't call method on interface until it's been cast properly.

	var buf bytes.Buffer
	werkerTask := &WerkerTask{}
	dec := gob.NewDecoder(&buf)
	err := dec.Decode(werkerTask)
	if err != nil {
		ocelog.IncludeErrField(err).Error()
		return nil
	}

	go w.build(werkerTask)
	return nil
}

// watchForResults sends the *Transport object over the transport channel for stream functions to process
func (w *WorkerMsgHandler) WatchForResults(hash string) {
	ocelog.Log().Debugf("adding hash ( %s ) & infochan to transport channel", hash)
	transport := &Transport{Hash: hash, InfoChan: w.infochan,}
	w.ChanChan <- transport
}

// build contains the logic for actually building. switches on type of proto message that was sent
// over the nsq queue
func (w *WorkerMsgHandler) build(werk *WerkerTask) {
	ocelog.Log().Debug("hash build ", werk.CheckoutHash)
	w.WatchForResults(werk.CheckoutHash)
	pipe, err := server.NewPipeline(nil, werk.Pipe)
	if err != nil {
		ocelog.IncludeErrField(err).Error("error building new pipeline")
	}
	quit := make(chan int8)
	done := make(chan int8)
	pipe.Run(quit, done)

	switch w.WerkConf.werkerType {
	case Docker:
		dockerPipe := pipe.JobsMap[werk.CheckoutHash]
		buildOutput, err := dockerPipe.Logs(true, true, true)
		defer buildOutput.Close()

		rd := bufio.NewReader(buildOutput)

		for {
			str, err := rd.ReadString('\n')
			if err != nil {
				log.Fatal("Read Error:", err)
				return
			}
			w.infochan <- []byte(str)
		}

		if err != nil {
			ocelog.IncludeErrField(err)
			return
		}
	case Kubernetes:
		//TODO: wait for kubernetes client
	}

	<-done
	ocelog.Log().Debugf("finished building id %s", werk.CheckoutHash)
}
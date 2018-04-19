package nsqwatch

import (
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	cred "bitbucket.org/level11consulting/ocelot/common/credentials"
	"bitbucket.org/level11consulting/ocelot/storage"
	"time"
)

// NsqWatch is for keeping an eye on Ocelot's dependencies and pausing the reception of messages
//  from NSQ if any other dependency goes down. the `paused` bool will be switched to true if the queue is shut down temporarily.
type NsqWatch struct {
	interval 	int64
	pConsumers  []*nsqpb.ProtoConsume
	remoteConf  cred.HealthyMaintainer
	store       storage.HealthyChkr
	paused      bool
}

func WatchAndPause(interval int64, consumers []*nsqpb.ProtoConsume, rc cred.CVRemoteConfig, store storage.OcelotStorage) {
	nsqwatch := &NsqWatch{
		interval: interval,
		pConsumers: consumers,
		remoteConf: rc,
		store: store,
	}
	nsqwatch.MaintainHealths()
}

// MaintainHealths will
func (nq *NsqWatch) MaintainHealths() {
	for {
		time.Sleep(time.Second * time.Duration(nq.interval))
		switch nq.paused {
		case false:
			if !nq.store.Healthy() || !nq.remoteConf.Healthy() {
				ocelog.Log().Error("DEPENDENCIES ARE DOWN!! PAUSING NSQ FLOW!")
				for _, protoConsumer := range nq.pConsumers {
					protoConsumer.Pause()
				}
				nq.paused = true
			}
		case true:
			err := nq.remoteConf.Reconnect()
			if err != nil {
				ocelog.IncludeErrField(err).Error("could not reconnect to remote config")
				continue
			}
			if nq.store.Healthy() {
				ocelog.Log().Info("services are back up! un-pausing nsq flow!")
				for _, protoConsumer := range nq.pConsumers {
					protoConsumer.UnPause()
				}
				nq.paused = false
				continue
			}
			ocelog.Log().Error("could not reconnect to database")
		}
	}
}
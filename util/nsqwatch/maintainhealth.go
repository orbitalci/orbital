package nsqwatch

import (
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"time"
)

type NsqWatch struct {
	interval 	int64
	pConsumers  []*nsqpb.ProtoConsume
	remoteConf  cred.CVRemoteConfig
	store       storage.OcelotStorage
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

func (nq *NsqWatch) MaintainHealths() {
	var paused bool
	for {
		time.Sleep(time.Second * time.Duration(nq.interval))
		switch paused {
		case false:
			if !nq.store.Healthy() || !nq.remoteConf.Healthy() {
				ocelog.Log().Error("DEPENDENCIES ARE DOWN!! PAUSING NSQ FLOW!")
				for _, protoConsumer := range nq.pConsumers {
					protoConsumer.Pause()
				}
				paused = true
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
				paused = false
				continue
			}
			ocelog.Log().Error("could not reconnect to database")
		}
	}
}
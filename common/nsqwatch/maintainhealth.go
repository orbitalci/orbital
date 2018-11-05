package nsqwatch

import (
	"github.com/prometheus/client_golang/prometheus"
	ocelog "github.com/shankj3/go-til/log"
	"github.com/shankj3/go-til/nsqpb"
	cred "github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/common/du"
	"github.com/shankj3/ocelot/storage"
	"time"
)

var (
	paused = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ocelot_werker_paused",
		Help: "integer boolean for whether or not healthy maintainer has paused the nsq flow",
	}, []string{"failed_dependency"})
)

func resetPaused() {
	paused.WithLabelValues("remoteConfig").Set(0)
	paused.WithLabelValues("store").Set(0)
	paused.WithLabelValues("diskUtility").Set(0)
}

func init() {
	prometheus.MustRegister(paused)
	resetPaused()
}

// NsqWatch is for keeping an eye on Ocelot's dependencies and pausing the reception of messages
//  from NSQ if any other dependency goes down. the `paused` bool will be switched to true if the queue is shut down temporarily.
type NsqWatch struct {
	interval   int64
	pConsumers []*nsqpb.ProtoConsume
	remoteConf cred.HealthyMaintainer
	store      storage.HealthyChkr
	paused     bool
	diskUtilityCheck *du.HCer
}

func WatchAndPause(interval int64, consumers []*nsqpb.ProtoConsume, rc cred.CVRemoteConfig, store storage.OcelotStorage, duHealthChecker *du.HCer) {
	nsqwatch := &NsqWatch{
		interval:         interval,
		pConsumers:       consumers,
		remoteConf:       rc,
		store:            store,
		diskUtilityCheck: duHealthChecker,
	}
	nsqwatch.MaintainHealths()
}

// MaintainHealths will
func (nq *NsqWatch) MaintainHealths() {
	for {
		time.Sleep(time.Second * time.Duration(10))
		switch nq.paused {
		case false:
			var pause bool
			if !nq.store.Healthy() {
				ocelog.Log().Error("STORAGE IS DOWN!")
				paused.WithLabelValues("store").Set(1)
				pause = true
			}
			if !nq.remoteConf.Healthy() {
				ocelog.Log().Error("REMOTE CONFIG IS DOWN!")
				paused.WithLabelValues("remoteConfig").Set(1)
				pause = true
			}
			if err := nq.diskUtilityCheck.Healthy(); err != nil {
				ocelog.Log().WithField("diskUtilityError", err).Error("DISK USAGE IS NO GOOD!")
				paused.WithLabelValues("diskUtility").Set(1)
				pause = true
			}
			if pause == true {
				ocelog.Log().Error("DEPENDENCIES ARE DOWN!! PAUSING NSQ FLOW!")
				for _, protoConsumer := range nq.pConsumers {
					protoConsumer.Pause()
				}
				nq.paused = true
			}
		case true:
			rcErr := nq.remoteConf.Reconnect()
			storeOk := nq.store.Healthy()
			duErr := nq.diskUtilityCheck.Healthy()
			if rcErr != nil || !storeOk || duErr != nil {
				ocelog.Log().WithField("remoteConfigError", rcErr).WithField("storeOk", storeOk).WithField("diskUtilityError", duErr).Error("dependencies are still down! not unpausing NSQ flow!")
			} else {
				ocelog.Log().Info("services are back up! un-pausing nsq flow!")
				for _, protoConsumer := range nq.pConsumers {
					protoConsumer.UnPause()
				}
				nq.paused = false
				resetPaused()
			}
		}
	}
}

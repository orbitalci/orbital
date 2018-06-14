package launcher

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shankj3/ocelot/build/basher"
	"github.com/shankj3/ocelot/build/integrations"
	"github.com/shankj3/ocelot/build/valet"
	"github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/storage"
)

// create a struct that is werker msg handler creates when it receives a new nsq message

type launcher struct {
	*models.WerkerFacts
	RemoteConf   credentials.CVRemoteConfig
	infochan     chan []byte
	StreamChan   chan *models.Transport
	BuildCtxChan chan *models.BuildContext
	Basher       *basher.Basher
	Store        storage.OcelotStorage
	BuildValet   *valet.Valet
	handler      models.VCSHandler
	integrations []integrations.StringIntegrator
}

func NewLauncher(facts *models.WerkerFacts,
	remoteConf credentials.CVRemoteConfig,
	streamChan chan *models.Transport,
	BuildCtxChan chan *models.BuildContext,
	bshr *basher.Basher,
	store storage.OcelotStorage,
	bv *valet.Valet) *launcher {
	return &launcher{
		WerkerFacts:  facts,
		RemoteConf:   remoteConf,
		StreamChan:   streamChan,
		BuildCtxChan: BuildCtxChan,
		Basher:       bshr,
		Store:        store,
		BuildValet:   bv,
		infochan:     make(chan []byte),
		integrations: getIntegrationList(),
	}
}

/// metrics
//var jobsInQueue = prometheus.NewGauge(
//	prometheus.GaugeOpts{
//		Name: "jobs_in_queue",
//		Help: "Current number of jobs in the queue",
//	},
//)
var (
	activeBuilds = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_builds",
			Help: "Number of builds currently in progress",
		},
	)
)
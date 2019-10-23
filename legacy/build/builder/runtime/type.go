package runtime

import (
	"github.com/level11consulting/orbitalci/build/builder/shell"
	"github.com/level11consulting/orbitalci/build/integrations"
	"github.com/level11consulting/orbitalci/build/buildmonitor"
	"github.com/level11consulting/orbitalci/models"
	"github.com/level11consulting/orbitalci/server/config"
	"github.com/level11consulting/orbitalci/storage"
)

// create a struct that is werker msg handler creates when it receives a new nsq message

type launcher struct {
	*models.WerkerFacts
	RemoteConf   config.CVRemoteConfig
	infochan     chan []byte
	StreamChan   chan *models.Transport
	BuildCtxChan chan *models.BuildContext
	Basher       *shell.Basher
	Store        storage.OcelotStorage
	BuildMonitor   *buildmonitor.BuildMonitor
	handler      models.VCSHandler
	integrations []integrations.StringIntegrator
	binaryIntegs []integrations.BinaryIntegrator
}

func NewLauncher(facts *models.WerkerFacts,
	remoteConf config.CVRemoteConfig,
	streamChan chan *models.Transport,
	BuildCtxChan chan *models.BuildContext,
	bshr *shell.Basher,
	store storage.OcelotStorage,
	bv *buildmonitor.BuildMonitor) *launcher {
	return &launcher{
		WerkerFacts:  facts,
		RemoteConf:   remoteConf,
		StreamChan:   streamChan,
		BuildCtxChan: BuildCtxChan,
		Basher:       bshr,
		Store:        store,
		BuildMonitor:   bv,
		infochan:     make(chan []byte),
		integrations: getIntegrationList(),
		binaryIntegs: getBinaryIntegList(bshr.LoopbackIp, facts.ServicePort),
	}
}

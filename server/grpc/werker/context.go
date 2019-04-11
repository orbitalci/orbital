package werker

import (
	"github.com/level11consulting/ocelot/build/streaminglogs"
	"github.com/level11consulting/ocelot/build/valet"
	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/storage"
	consulet "github.com/shankj3/go-til/consul"
	ocelog "github.com/shankj3/go-til/log"
)

type WerkerContext struct {
	*models.WerkerFacts
	consul     *consulet.Consulet
	store      storage.OcelotStorage
	streamPack *streaminglogs.StreamPack
	killValet  *valet.ContextValet
}

func getWerkerContext(conf *models.WerkerFacts, store storage.OcelotStorage, contextValet *valet.ContextValet) *WerkerContext {
	werkerConsul, err := consulet.Default()
	if err != nil {
		ocelog.IncludeErrField(err)
	}
	werkerCtx := &WerkerContext{
		WerkerFacts: conf,
		consul:      werkerConsul,
		killValet:   contextValet,
		store:       store,
	}
	return werkerCtx
}

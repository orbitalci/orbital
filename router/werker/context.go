package werker

import (
	//"encoding/json"
	//"net/http"
	//"time"

	consulet "github.com/shankj3/go-til/consul"
	ocelog "github.com/shankj3/go-til/log"
	"github.com/shankj3/ocelot/build/streamer"
	"github.com/shankj3/ocelot/build/valet"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/storage"
)

type WerkerContext struct {
	*models.WerkerFacts
	consul     *consulet.Consulet
	store      storage.OcelotStorage
	streamPack *streamer.StreamPack
	killValet  *valet.ContextValet
}

//
//func (w *WerkerContext) dumpData(wr http.ResponseWriter, r *http.Request) {
//	ocelog.Log().Info("writing out data for buildInfo")
//	wr.Header().Set("content-type", "application/json")
//	dataMap := make(map[string]int)
//	dataMap["time"] = int(time.Now().Unix())
//	wr.WriteHeader(http.StatusOK)
//	for hash, bytearray := range w.buildInfo {
//		dataMap[hash] = len(bytearray.GetData())
//	}
//	bit, err := json.Marshal(dataMap)
//	if err != nil {
//		ocelog.IncludeErrField(err).Error("couldn't marshal for dump")
//		wr.WriteHeader(http.StatusInternalServerError)
//		return
//	}
//	wr.Write(bit)
//}

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

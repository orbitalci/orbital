package hookhandler

import (
	"bitbucket.org/level11consulting/go-til/deserialize"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/client/validate"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"github.com/golang/protobuf/proto"
	"fmt"
)

type BuildHookHandler struct {
	RemoteConfig cred.CVRemoteConfig
	Deserializer *deserialize.Deserializer
	Validator    *validate.OcelotValidator
	Producer     *nsqpb.PbProduce
}

// UnmarshalAndProcess is called by the nsq consumer to handle the build message
func (b *BuildHookHandler) UnmarshalAndProcess(msg []byte, done chan int) error {
	//ocelog.Log().Debug("unmarshaling build obj and processing")
	buildTask := &models.AcctRepoAndHash{}
	if err := proto.Unmarshal(msg, buildTask); err != nil {
		ocelog.IncludeErrField(err).Warning("unmarshal error")
		return err
	}

	store, err := b.RemoteConfig.GetOcelotStorage()
	if err != nil {
		ocelog.IncludeErrField(err).Warning("could not get storage")
		return err
	}

	buildSum, err := store.RetrieveLatestSum(buildTask.PartialHash)
	if err != nil {
		ocelog.IncludeErrField(err).Warning(fmt.Sprintf("could not retrieve latest sum for %s", buildTask.PartialHash))
		return err
	}

	buildConf, token, err := GetBBConfig(b.RemoteConfig, buildTask.AcctRepo, buildSum.Hash, b.Deserializer)
	if err != nil {
		ocelog.IncludeErrField(err).Warning(fmt.Sprintf("could not retrieve build configuration for for %s", buildTask.PartialHash))
		return err
	}

	if err = QueueAndStore(buildSum.Hash, buildSum.Branch, buildTask.AcctRepo, token, b.RemoteConfig, buildConf, b.Validator, b.Producer, store); err != nil {
		ocelog.IncludeErrField(err).Warning(fmt.Sprintf("could not queue up build for %s", buildTask.PartialHash))
		return err
	}

	done <- 1
	return nil
}

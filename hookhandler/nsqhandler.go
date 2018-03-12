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
func (b *BuildHookHandler) UnmarshalAndProcess(msg []byte, done chan int, finish chan int) error {
	defer func(){
		if r := recover(); r != nil {
			// add to finish channel so that we don't requeue
			finish <- 1
			ocelog.Log().Fatal("a panic occurred, exiting: ", r)
		}
	}()
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

	fullHash := buildTask.PartialHash

	buildSum, err := store.RetrieveLatestSum(buildTask.PartialHash)
	if err != nil { //continue after error because maybe they're passing full hash that's just not in our db yet
		ocelog.IncludeErrField(err).Warning(fmt.Sprintf("there is no build matching hash %s in db", buildTask.PartialHash))
	} else {
		fullHash = buildSum.Hash
	}

	buildConf, token, err := GetBBConfig(b.RemoteConfig, buildTask.AcctRepo, fullHash, b.Deserializer)
	if err != nil {
		ocelog.IncludeErrField(err).Warning(fmt.Sprintf("could not retrieve build configuration for for %s", buildTask.PartialHash))
		return err
	}

	if err = QueueAndStore(fullHash, buildSum.Branch, buildTask.AcctRepo, token, b.RemoteConfig, buildConf, b.Validator, b.Producer, store); err != nil {
		ocelog.IncludeErrField(err).Warning(fmt.Sprintf("could not queue up build for %s", buildTask.PartialHash))
		return err
	}

	done <- 1
	return nil
}

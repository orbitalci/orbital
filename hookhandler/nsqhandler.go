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
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"time"
	"github.com/hashicorp/vault/command"
	"bitbucket.org/level11consulting/ocelot/util/storage"
)

type BuildHookHandler struct {
	RemoteConfig cred.CVRemoteConfig
	Deserializer *deserialize.Deserializer
	Validator    *validate.OcelotValidator
	Producer     *nsqpb.PbProduce
}

// UnmarshalAndProcess is called by the nsq consumer to handle the build message
func (b *BuildHookHandler) UnmarshalAndProcess(msg []byte, done chan int) error {
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

	var stageMessages []string
	var buildId int64
	var fullHash string

	acct, repo := getAcctRepo(buildTask.AcctRepo)
	buildSum, err := store.RetrieveLatestSum(buildTask.PartialHash)

	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); !ok {
			ocelog.IncludeErrField(err).Error("a serious error has occurred while performing lookup for latest sum starting with " + buildTask.PartialHash)
			return err
		}
		//at this point error must be because we couldn't find hash starting with query
		ocelog.IncludeErrField(err).Warning(fmt.Sprintf("there is no build starting with hash %s in db", buildTask.PartialHash))
		stageMessages = append(stageMessages, fmt.Sprintf("there is no build starting with %s in db", buildTask.PartialHash))


	} else {
		fullHash = buildSum.Hash
		stageMessages = append(stageMessages, fmt.Sprintf("found a preexiting build with hash %s belonging to %s/%s", fullHash, acct, repo))
		buildId = storeSummaryToDb(store, fullHash, repo, "", acct)
		if buildId > 0 {
			stageResult := getHookhandlerStageResult(buildId)
			stageMessages = append(stageMessages, fmt.Sprintf("starting build %d for %s/%s with hash %s", buildId, acct, repo, fullHash))
			err = storeStageToDb(store, stageResult)
			if err != nil {
				ocelog.IncludeErrField(err).Warning(fmt.Sprintf("there is no build matching hash %s in db", buildTask.PartialHash))
			}
		} else {
			ocelog.IncludeErrField(err).Warning(fmt.Sprintf("for some reason we could not get a build id for %s/%s", acct, repo))
		}
	}

	buildConf, token, err := GetBBConfig(b.RemoteConfig, buildTask.AcctRepo, fullHash, b.Deserializer)
	if err != nil {
		ocelog.IncludeErrField(err).Warning(fmt.Sprintf("could not retrieve build configuration for for %s", buildTask.PartialHash))
		return err
	}


	//TODO: we can't call queue and store here because we're already keeping track of our own list of messages to store for stages
	//if err = QueueAndStore(fullHash, buildSum.Branch, buildTask.AcctRepo, token, b.RemoteConfig, buildConf, b.Validator, b.Producer, store); err != nil {
	//	ocelog.IncludeErrField(err).Warning(fmt.Sprintf("could not queue up build for %s", buildTask.PartialHash))
	//	return err
	//}

	done <- 1
	return nil
}

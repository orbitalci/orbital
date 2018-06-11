package build_signaler

import (
	"github.com/shankj3/go-til/deserialize"
	"github.com/shankj3/go-til/log"
	"github.com/shankj3/go-til/nsqpb"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"

	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/common"
	"github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/storage"
)

func NewSignaler(RC credentials.CVRemoteConfig, dese *deserialize.Deserializer, producer nsqpb.Producer, ocyValidator *build.OcelotValidator, store storage.OcelotStorage) *Signaler {
	return &Signaler{
		RC: RC,
		Deserializer: dese,
		Producer: producer,
		OcyValidator: ocyValidator,
		Store: store,
	}
}

type Signaler struct {
	RC credentials.CVRemoteConfig
	*deserialize.Deserializer
	Producer     nsqpb.Producer
	OcyValidator *build.OcelotValidator
	Store        storage.OcelotStorage
}

// CheckViableThenQueueAndStore is a dumb name, but i can't think of a better one. it will first
//- check if the build is "viable", ie if it is in the accepted branches list and none of the commits contain a skip message. if it isn't, it won't queue and will return a NotViable error
//- will then run queueAndStore, which will:
//  - check if build in consul, if it is it will not add to queue and return a NotViable error
//  - if the above doesn't return an error....
//  - store the initial summary in the database
//  - validate that the configuration is good
func (s *Signaler) CheckViableThenQueueAndStore(hash, authToken, branch, acctRepo string, buildConf *pb.BuildConfig, commits []*pb.Commit, force bool) error {
	if queueError := s.OcyValidator.ValidateViability(branch, buildConf.Branches, commits, force); queueError != nil {
		log.IncludeErrField(queueError).Info("not queuing! this is fine, just doesn't fit requirements")
		return queueError
	}
	return s.queueAndStore(hash, authToken, branch, acctRepo, buildConf)
}

// queueAndStore is the muscle; it will:
//  - check if build in consul, if it is it will not add to queue and return a NotViable error
//  - if the above doesn't return an error....
//  - store the initial summary in the database
//  - validate that the configuration is good and add to the queue. If the build configuration (ocelot.yml) is not good, then it will store in the database that it failed validation along with the reason why
func (s *Signaler) queueAndStore(hash, authToken, branch, acctRepo string, buildConf *pb.BuildConfig) error {
	log.Log().Debug("Storing initial results in db")
	account, repo, err := common.GetAcctRepo(acctRepo)
	if err != nil {
		return err
	}
	alreadyBuilding, err := build.CheckBuildInConsul(s.RC.GetConsul(), hash)
	if alreadyBuilding {
		log.Log().Info("kindly refusing to add to queue because this hash is already building")
		return build.NoViability("this hash is already building in ocelot, therefore not adding to queue")
	}
	// tell the database (and therefore all of ocelot) that this build is a-happening. or at least that it exists.
	id, err := storeSummaryToDb(s.Store, hash, repo, branch, account)
	if err != nil {
		return err
	}

	sr := getSignalerStageResult(id)
	if err = s.validateAndQueue(buildConf, sr, branch, hash, authToken, acctRepo, id); err != nil {
		// we do want to add a runtime here
		err = s.Store.StoreFailedValidation(id)
		if err != nil {
			log.IncludeErrField(err).Error("unable to update summary!")
		}
		// we dont' want to return here, cuz then it doesn't store
		// unless its supposed to be saving somewhere else?
		// return err
	} else {
		storeQueued(s.Store, id)
	}
	if err := storeStageToDb(s.Store, sr); err != nil {
		log.IncludeErrField(err).Error("unable to add hookhandler stage details")
		return err
	}
	return nil
}

// validateAndQueue will use the OcyValidator to make sure that the config is up to spec, if it is then it will add it to the build queue.
//  If it isn't, it will store a FAILED VALIDATION with the validation errors.
func (s *Signaler) validateAndQueue(buildConf *pb.BuildConfig, sr *models.StageResult, branch, hash, authToken, acctRepo string, buildId int64) error {
	if err := s.OcyValidator.ValidateConfig(buildConf, nil); err == nil {
		s.tellWerker(buildConf, hash, authToken, buildId, branch, acctRepo)
		log.Log().Debug("told werker!")
		PopulateStageResult(sr, 0, "Passed initial validation "+models.CHECKMARK, "")
	} else {
		PopulateStageResult(sr, 1, "Failed initial validation", err.Error())
		return err
	}
	return nil
}


//tellWerker is a private helper function for building a werker task and giving it to nsq
func (s *Signaler) tellWerker(buildConf *pb.BuildConfig,
	hash string,
	authToken string,
	dbid int64,
	branch string,
	acctRepo string) {
	// get one-time token use for access to vault
	token, err := s.RC.GetVault().CreateThrowawayToken()
	if err != nil {
		log.IncludeErrField(err).Error("unable to create one-time vault token")
		return
	}
	werkerTask := &pb.WerkerTask{
		VaultToken:   token,
		CheckoutHash: hash,
		Branch:       branch,
		BuildConf:    buildConf,
		VcsToken:     authToken,
		VcsType:      pb.SubCredType_BITBUCKET,
		FullName:     acctRepo,
		Id:           dbid,
	}

	go s.Producer.WriteProto(werkerTask, build.DetermineTopic(buildConf.MachineTag))
}



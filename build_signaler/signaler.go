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
func (s *Signaler) CheckViableThenQueueAndStore(hash, authToken, branch, acctRepo string, buildConf *pb.BuildConfig, commits []*pb.Commit, force bool, sigType pb.SignaledBy, prData *pb.PullRequest) error {
	if queueError := s.OcyValidator.ValidateViability(branch, buildConf.Branches, commits, force, prData); queueError != nil {
		log.IncludeErrField(queueError).Info("not queuing! this is fine, just doesn't fit requirements")
		return queueError
	}
	task := buildInitialWerkerTask(buildConf, hash, authToken, branch, acctRepo, sigType, prData)
	return s.queueAndStore(task)
}

// queueAndStore is the muscle; it will:
//  - check if build in consul, if it is it will not add to queue and return a NotViable error
//  - if the above doesn't return an error....
//  - store the initial summary in the database
//  - validate that the configuration is good and add to the queue. If the build configuration (ocelot.yml) is not good, then it will store in the database that it failed validation along with the reason why
func (s *Signaler) queueAndStore(task *pb.WerkerTask) error {
	log.Log().Debug("Storing initial results in db")
	account, repo, err := common.GetAcctRepo(task.FullName)
	if err != nil {
		return err
	}
	alreadyBuilding, err := build.CheckBuildInConsul(s.RC.GetConsul(), task.CheckoutHash)
	if alreadyBuilding {
		log.Log().Info("kindly refusing to add to queue because this hash is already building")
		return build.NoViability("this hash is already building in ocelot, therefore not adding to queue")
	}
	// tell the database (and therefore all of ocelot) that this build is a-happening. or at least that it exists.
	id, err := storeSummaryToDb(s.Store, task.CheckoutHash, repo, task.Branch, account)
	if err != nil {
		return err
	}

	sr := getSignalerStageResult(id)
	vaultToken, _ := s.RC.GetVault().CreateThrowawayToken()
	updateWerkerTask(task, id, vaultToken)
	if err = s.validateAndQueue(task, sr); err != nil {
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
func (s *Signaler) validateAndQueue(task *pb.WerkerTask, sr *models.StageResult) error {
	if err := s.OcyValidator.ValidateConfig(task.BuildConf, nil); err == nil {
		s.Producer.WriteProto(task, build.DetermineTopic(task.BuildConf.MachineTag))
		log.Log().Debug("told werker!")
		PopulateStageResult(sr, 0, "Passed initial validation "+models.CHECKMARK, "")
	} else {
		PopulateStageResult(sr, 1, "Failed initial validation", err.Error())
		return err
	}
	return nil
}

//buildInitialWerkerTask will create a WerkerTask object with all the fields taht are not reliant on a database transaction, basically everyhting we know right when we know we want to try and queue a build
func buildInitialWerkerTask(buildConf *pb.BuildConfig,
	hash string,
	authToken string,
	branch string,
	acctRepo string,
	sigType pb.SignaledBy,
	prData *pb.PullRequest) *pb.WerkerTask {
	return &pb.WerkerTask{
		CheckoutHash:  hash,
		Branch:        branch,
		BuildConf:     buildConf,
		VcsToken:      authToken,
		VcsType:       pb.SubCredType_BITBUCKET,
		FullName:      acctRepo,
		SignaledBy:    sigType,
		Pr:            prData,
	}
}

//updateWerkerTask will fill in the fields of everything generated after, ie the build id from the database insertion and the vault token
func updateWerkerTask(task *pb.WerkerTask, dbId int64, token string) {
	task.Id = dbId
	task.VaultToken = token
}
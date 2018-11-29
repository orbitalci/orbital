package build_signaler

import (
	"github.com/pkg/errors"
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
		RC:           RC,
		Deserializer: dese,
		Producer:     producer,
		OcyValidator: ocyValidator,
		Store:        store,
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
//- check if the build is "viable", ie if it is in the accepted branches list and none of the commits contain a skip Message. if it isn't, it won't queue and will return a NotViable error
//- will then run QueueAndStore, which will:
//  - check if build in consul, if it is it will not add to queue and return a NotViable error
//  - if the above doesn't return an error....
//  - store the initial summary in the database
//  - validate that the configuration is good
func (s *Signaler) CheckViableThenQueueAndStore(task *pb.WerkerTask, force bool, commits []*pb.Commit) error {
	if queueError := s.OcyValidator.ValidateViability(task.Branch, task.BuildConf.Branches, commits, force); queueError != nil {
		log.IncludeErrField(queueError).Info("not queuing! this is fine, just doesn't fit requirements")
		return queueError
	}
	return s.QueueAndStore(task)
}

// QueueAndStore is the muscle; it will:
//  - check if build in consul, if it is it will not add to queue and return a NotViable error
//  - if the above doesn't return an error....
//  - store the initial summary in the database
//  - validate that the configuration is good and add to the queue. If the build configuration (ocelot.yml) is not good, then it will store in the database that it failed validation along with the reason why
func (s *Signaler) QueueAndStore(task *pb.WerkerTask) error {
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
	//get vcs cred to attach the database id to the build summary
	vcsCred, err := credentials.GetVcsCreds(s.Store, task.FullName, s.RC, task.VcsType)
	if err != nil {
		log.IncludeErrField(err).Error("unable to get vcs creds")
		return errors.Wrap(err, "unable to retrieve vcs creds")
	}
	// tell the database (and therefore all of ocelot) that this build is a-happening. or at least that it exists.
	id, err := storeSummaryToDb(s.Store, task.CheckoutHash, repo, task.Branch, account, task.SignaledBy, vcsCred.GetId())
	if err != nil {
		return err
	}
	sr := getSignalerStageResult(id)

	// after storing that this build was recived, check to make sure the build config is even worthy of our time
	if err := s.OcyValidator.ValidateConfig(task.BuildConf, nil); err != nil {
		PopulateStageResult(sr, 1, "Failed initial validation", err.Error())
		err = s.Store.StoreFailedValidation(id)
		if err != nil {
			log.IncludeErrField(err).Error("unable to update summary!")
		}
	} else {
		PopulateStageResult(sr, 0, "Passed initial validation "+models.CHECKMARK, "")
		vaultToken, _ := s.RC.GetVault().CreateThrowawayToken()
		updateWerkerTask(task, id, vaultToken)
		if err = s.Producer.WriteProto(task, build.DetermineTopic(task.BuildConf.MachineTag)); err != nil {
			log.IncludeErrField(err).WithField("buildId", task.Id).Error("error writing proto msg for build")
		}
		if err = storeQueued(s.Store, id); err != nil {
			log.IncludeErrField(err).WithField("buildId", task.Id).Error("error storing queued state")
		}
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
		if err := s.Producer.WriteProto(task, build.DetermineTopic(task.BuildConf.MachineTag)); err != nil {
			log.IncludeErrField(err).WithField("buildId", task.Id).Error("error writing proto msg for build")
		} else {
			log.Log().Debug("told werker!")
		}
		PopulateStageResult(sr, 0, "Passed initial validation "+models.CHECKMARK, "")
	} else {
		PopulateStageResult(sr, 1, "Failed initial validation", err.Error())
		return err
	}
	return nil
}

//BuildInitialWerkerTask will create a WerkerTask object with all the fields taht are not reliant on a database transaction, basically everyhting we know right when we know we want to try and queue a build
//todo: take out the hard coded VcsType of bitbucket
func BuildInitialWerkerTask(buildConf *pb.BuildConfig,
	hash string,
	authToken string,
	branch string,
	acctRepo string,
	sigType pb.SignaledBy,
	prData *pb.PrWerkerData,
	vcsType pb.SubCredType) *pb.WerkerTask {
	return &pb.WerkerTask{
		CheckoutHash: hash,
		Branch:       branch,
		BuildConf:    buildConf,
		VcsToken:     authToken,
		VcsType:      vcsType,
		FullName:     acctRepo,
		SignaledBy:   sigType,
		PrData:       prData,
	}
}

//updateWerkerTask will fill in the fields of everything generated after, ie the build id from the database insertion and the vault token
func updateWerkerTask(task *pb.WerkerTask, dbId int64, token string) {
	task.Id = dbId
	task.VaultToken = token
}

package build_signaler

import (
	"errors"
	"fmt"
	"time"

	"github.com/shankj3/go-til/deserialize"
	"github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"
	"github.com/shankj3/go-til/nsqpb"
	ocevault "github.com/shankj3/go-til/vault"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/common"
	cred "github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"
)

//QueueAndStore will create a werker task and put it on the queue, then update database
func QueueAndStore(hash, branch, accountRepo, bbToken string,
	remoteConfig cred.CVRemoteConfig,
	buildConf *pb.BuildConfig,
	validator *build.OcelotValidator,
	producer *nsqpb.PbProduce,
	store storage.OcelotStorage) error {

	log.Log().Debug("Storing initial results in db")
	account, repo, err := common.GetAcctRepo(accountRepo)
	if err != nil {
		return err
	}
	vaulty := remoteConfig.GetVault()
	consul := remoteConfig.GetConsul()
	alreadyBuilding, err := build.CheckBuildInConsul(consul, hash)
	if alreadyBuilding {
		log.Log().Info("kindly refusing to add to queue because this hash is already building")
		return errors.New("this hash is already building in ocelot, therefore not adding to queue")
	}

	id, err := storeSummaryToDb(store, hash, repo, branch, account)
	if err != nil {
		return err
	}

	sr := getSignalerStageResult(id)
	// stageResult.BuildId, stageResult.Stage, stageResult.Error, stageResult.StartTime, stageResult.StageDuration, stageResult.Status, stageResult.Messages
	if err = ValidateAndQueue(buildConf, branch, validator, vaulty, producer, sr, id, hash, accountRepo, bbToken); err != nil {
		// we do want to add a runtime here
		err = store.StoreFailedValidation(id)
		if err != nil {
			log.IncludeErrField(err).Error("unable to update summary!")
		}
		// we dont' want to return here, cuz then it doesn't store
		// unless its supposed to be saving somewhere else?
		// return err
	} else {
		storeQueued(store, id)
	}
	if err := storeStageToDb(store, sr); err != nil {
		log.IncludeErrField(err).Error("unable to add hookhandler stage details")
		return err
	}
	return nil
}

func storeStageToDb(store storage.BuildStage, stageResult *models.StageResult) error {
	if err := store.AddStageDetail(stageResult); err != nil {
		log.IncludeErrField(err).Error("unable to store hookhandler stage details to db")
		return err
	}
	return nil
}
func storeQueued(store storage.BuildSum, id int64) error {
	err := store.SetQueueTime(id)
	if err != nil {
		log.IncludeErrField(err).Error("unable to update queue time in build summary table")
	}
	return err
}

func storeSummaryToDb(store storage.BuildSum, hash, repo, branch, account string) (int64, error) {
	id, err := store.AddSumStart(hash, account, repo, branch)
	if err != nil {
		log.IncludeErrField(err).Error("unable to store summary details to db")
		return 0, err
	}
	return id, nil
}

//todo: pull out check for vcsHandler == nil logic, then this can be just GetConfig()
// todo (cont): write something in remote to switch between subtypes to instantiate the correct VCSHandler implementation
//GetConfig returns the protobuf ocelot.yaml, a valid bitbucket token belonging to that repo, and possible err.
//If a VcsHandler is passed, this method will use the existing handler to retrieve the bb config. In that case,
//***IT WILL NOT RETURN A VALID TOKEN FOR YOU - ONLY BUILD CONFIG***
func GetConfig(repoFullName string, checkoutCommit string, deserializer *deserialize.Deserializer, vcsHandler models.VCSHandler) (*pb.BuildConfig, error) {
	//var bbHandler remote.VCSHandler
	//var token string

	if vcsHandler == nil {
		//cfg, err1 := cred.GetVcsCreds(store, repoFullName, remoteConfig)
		//if err1 != nil {
		//	log.IncludeErrField(err1).Error()
		//	return nil, "", err1
		//}
		//var err error
		//log.Log().WithField("identifier", cfg.GetIdentifier()).Infof("trying bitbucket config")
		//bbHandler, token, err = bitbucket.GetBitbucketClient(cfg)
		//if err != nil {
		//	log.IncludeErrField(err).Error()
		//	return nil, "", err
		return nil, errors.New("vcs handler cannot be nul")
	}

	fileBytz, err := vcsHandler.GetFile("ocelot.yml", repoFullName, checkoutCommit)
	if err != nil {
		log.IncludeErrField(err).Error()
		return nil, err
	}

	conf, err := CheckForBuildFile(fileBytz, deserializer)
	return conf, err
}

//CheckForBuildFile will try to retrieve an ocelot.yaml file for a repository and return the protobuf message
func CheckForBuildFile(buildFile []byte, deserializer *deserialize.Deserializer) (*pb.BuildConfig, error) {
	conf := &pb.BuildConfig{}
	fmt.Println(string(buildFile))
	if err := deserializer.YAMLToStruct(buildFile, conf); err != nil {
		if err != ocenet.FileNotFound {
			log.IncludeErrField(err).Error("unable to get build conf")
			return conf, err
		}
		log.Log().Debugf("no ocelot yml found")
		return conf, err
	}
	return conf, nil
}

//Validate is a util class that will validate your ocelot.yml + build config, queue the message to werker if
//it passes
func ValidateAndQueue(buildConf *pb.BuildConfig,
	branch string,
	validator *build.OcelotValidator,
	vaulty ocevault.Vaulty,
	producer *nsqpb.PbProduce,
	sr *models.StageResult,
	buildId int64,
	hash, fullAcctRepo, bbToken string) error {
	if err := validator.ValidateWithBranch(buildConf, branch, nil); err == nil {
		tellWerker(buildConf, vaulty, producer, hash, fullAcctRepo, bbToken, buildId, branch)
		log.Log().Debug("told werker!")
		PopulateStageResult(sr, 0, "Passed initial validation "+models.CHECKMARK, "")
	} else {
		PopulateStageResult(sr, 1, "Failed initial validation", err.Error())
		return err
	}
	return nil
}

//TellWerker is a private helper function for building a werker task and giving it to nsq
func tellWerker(buildConf *pb.BuildConfig,
	vaulty ocevault.Vaulty,
	producer *nsqpb.PbProduce,
	hash string,
	fullName string,
	bbToken string,
	dbid int64,
	branch string) {
	// get one-time token use for access to vault
	token, err := vaulty.CreateThrowawayToken()
	if err != nil {
		log.IncludeErrField(err).Error("unable to create one-time vault token")
		return
	}
	werkerTask := &pb.WerkerTask{
		VaultToken:   token,
		CheckoutHash: hash,
		Branch:       branch,
		BuildConf:    buildConf,
		VcsToken:     bbToken,
		VcsType:      pb.SubCredType_BITBUCKET,
		FullName:     fullName,
		Id:           dbid,
	}

	go producer.WriteProto(werkerTask, build.DetermineTopic(buildConf.MachineTag))
}

func PopulateStageResult(sr *models.StageResult, status int, lastMsg, errMsg string) {
	sr.Messages = append(sr.Messages, lastMsg)
	sr.Status = status
	sr.Error = errMsg
	sr.StageDuration = time.Now().Sub(sr.StartTime).Seconds()
}

// all this moved to build_signaler.go
func getSignalerStageResult(id int64) *models.StageResult {
	start := time.Now()
	return &models.StageResult{
		Messages:      []string{},
		BuildId:       id,
		Stage:         models.HOOKHANDLER_VALIDATION,
		StartTime:     start,
		StageDuration: -99.99,
	}
}

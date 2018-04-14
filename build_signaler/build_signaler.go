package build_signaler

import (
	"strings"

	"bitbucket.org/level11consulting/go-til/deserialize"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	"bitbucket.org/level11consulting/ocelot/util/buildruntime"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/handler"
)

//QueueAndStore will create a werker task and put it on the queue, then update database
func QueueAndStore(hash, branch, accountRepo, bbToken string,
	remoteConfig cred.CVRemoteConfig,
	buildConf *pb.BuildConfig,
	validator *OcelotValidator,
	producer *nsqpb.PbProduce,
	store storage.OcelotStorage) error {

	ocelog.Log().Debug("Storing initial results in db")
	account, repo, err := GetAcctRepo(accountRepo)
	if err != nil {
		return err
	}
	vaulty := remoteConfig.GetVault()
	consul := remoteConfig.GetConsul()
	alreadyBuilding, err := buildruntime.CheckBuildInConsul(consul, hash)
	if alreadyBuilding {
		ocelog.Log().Info("kindly refusing to add to queue because this hash is already building")
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
			ocelog.IncludeErrField(err).Error("unable to update summary!")
		}
		// we dont' want to return here, cuz then it doesn't store
		// unless its supposed to be saving somewhere else?
		// return err
	} else {
		storeQueued(store, id)
	}
	if err := storeStageToDb(store, sr); err != nil {
		ocelog.IncludeErrField(err).Error("unable to add hookhandler stage details")
		return err
	}
	return nil
}

func storeStageToDb(store storage.BuildStage, stageResult *smods.StageResult) error {
	if err := store.AddStageDetail(stageResult); err != nil {
		ocelog.IncludeErrField(err).Error("unable to store hookhandler stage details to db")
		return err
	}
	return nil
}
func storeQueued(store storage.BuildSum, id int64) error {
	err := store.SetQueueTime(id)
	if err != nil {
		ocelog.IncludeErrField(err).Error("unable to update queue time in build summary table")
	}
	return err
}


func storeSummaryToDb(store storage.BuildSum, hash, repo, branch, account string) (int64, error) {
	id, err := store.AddSumStart(hash, account, repo, branch)
	if err != nil {
		ocelog.IncludeErrField(err).Error("unable to store summary details to db")
		return 0, err
	}
	return id, nil
}


//GetBBConfig returns the protobuf ocelot.yaml, a valid bitbucket token belonging to that repo, and possible err.
//If a VcsHandler is passed, this method will use the existing handler to retrieve the bb config. In that case,
//***IT WILL NOT RETURN A VALID TOKEN FOR YOU - ONLY BUILD CONFIG***
func GetBBConfig(remoteConfig cred.CVRemoteConfig, store storage.CredTable, repoFullName string, checkoutCommit string, deserializer *deserialize.Deserializer, vcsHandler handler.VCSHandler) (*pb.BuildConfig, string, error) {
	var bbHandler handler.VCSHandler
	var token string

	if vcsHandler == nil {
		cfg, err1 := GetVcsCreds(store, repoFullName, remoteConfig)
		if err1 != nil {
			ocelog.IncludeErrField(err1).Error()
			return nil, "", err1
		}
		var err error
		ocelog.Log().WithField("identifier", cfg.GetIdentifier()).Infof("trying bitbucket config")
		bbHandler, token, err = handler.GetBitbucketClient(cfg)
		if err != nil {
			ocelog.IncludeErrField(err).Error()
			return nil, "", err
		}
	} else {
		bbHandler = vcsHandler
	}

	fileBytz, err := bbHandler.GetFile("ocelot.yml", repoFullName, checkoutCommit)
	if err != nil {
		ocelog.IncludeErrField(err).Error()
		return nil, token, err
	}

	conf, err := CheckForBuildFile(fileBytz, deserializer)
	return conf, token, err
}

//CheckForBuildFile will try to retrieve an ocelot.yaml file for a repository and return the protobuf message
func CheckForBuildFile(buildFile []byte, deserializer *deserialize.Deserializer) (*pb.BuildConfig, error) {
	conf := &pb.BuildConfig{}
	fmt.Println(string(buildFile))
	if err := deserializer.YAMLToStruct(buildFile, conf); err != nil {
		if err != ocenet.FileNotFound {
			ocelog.IncludeErrField(err).Error("unable to get build conf")
			return conf, err
		}
		ocelog.Log().Debugf("no ocelot yml found")
		return conf, err
	}
	return conf, nil
}

//Validate is a util class that will validate your ocelot.yml + build config, queue the message to werker if
//it passes
func ValidateAndQueue(buildConf *pb.BuildConfig,
	branch string,
	validator *OcelotValidator,
	vaulty ocevault.Vaulty,
	producer *nsqpb.PbProduce,
	sr *smods.StageResult,
	buildId int64,
	hash, fullAcctRepo, bbToken string) error {
	if err := validateBuild(buildConf, branch, validator); err == nil {
		tellWerker(buildConf, vaulty, producer, hash, fullAcctRepo, bbToken, buildId, branch)
		ocelog.Log().Debug("told werker!")
		PopulateStageResult(sr, 0, "Passed initial validation " + smods.CHECKMARK, "")
	} else {
		PopulateStageResult(sr, 1, "Failed initial validation", err.Error())
		return err
	}
	return nil
}

//tellWerker is a private helper function for building a werker task and giving it to nsq
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
		ocelog.IncludeErrField(err).Error("unable to create one-time vault token")
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

	go producer.WriteProto(werkerTask, "build")
}

//before we build pipeline config for werker, validate and make sure this is good candidate
// - check if commit branch matches with ocelot.yaml branch and validate
func validateBuild(buildConf *pb.BuildConfig, branch string, validator *OcelotValidator) error {
	err := validator.ValidateConfig(buildConf, nil)

	if err != nil {
		ocelog.IncludeErrField(err).Error("failed validation")
		return err
	}

	for _, buildBranch := range buildConf.Branches {
		if buildBranch == "ALL" || buildBranch == branch {
			return nil
		}
	}
	ocelog.Log().Errorf("build does not match any branches listed: %v", buildConf.Branches)
	return errors.New(fmt.Sprintf("build does not match any branches listed: %v", buildConf.Branches))
}



func PopulateStageResult(sr *smods.StageResult, status int, lastMsg, errMsg string) {
	sr.Messages = append(sr.Messages, lastMsg)
	sr.Status = status
	sr.Error = errMsg
	sr.StageDuration = time.Now().Sub(sr.StartTime).Seconds()
}

// all this moved to build_signaler.go
func getSignalerStageResult(id int64) *smods.StageResult {
	start := time.Now()
	return &smods.StageResult{
		Messages: 	   []string{},
		BuildId:       id,
		Stage:         smods.HOOKHANDLER_VALIDATION,
		StartTime:     start,
		StageDuration: -99.99,
	}
}





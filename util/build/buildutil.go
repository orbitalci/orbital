package build

import (
	"bitbucket.org/level11consulting/ocelot/util/buildruntime"
	"strings"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"fmt"
	"errors"
	"bitbucket.org/level11consulting/go-til/deserialize"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	ocevault "bitbucket.org/level11consulting/go-til/vault"
	smods "bitbucket.org/level11consulting/ocelot/util/storage/models"
	"bitbucket.org/level11consulting/ocelot/util/handler"
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"time"
)

// this file contains build util functions used by both hookhandler and admin
// hookhandler uses this when it receives a new build and admin uses this
// when it receives a build command via command line

// helper
func GetAcctRepo(fullName string) (acct string, repo string, err error) {
	list := strings.Split(fullName, "/")
	if len(list) != 2 {
		return "","", BadFormat("acctRepo needs to be in format acct/repo")
	}
	acct = list[0]
	repo = list[1]
	return
}

func PopulateStageResult(sr *smods.StageResult, status int, lastMsg, errMsg string) {
	sr.Messages = append(sr.Messages, lastMsg)
	sr.Status = status
	sr.Error = errMsg
	sr.StageDuration = time.Now().Sub(sr.StartTime).Seconds()
}

//get vcs creds will build a path for you based on the full name of the repo and return the vcsCredentials corresponding
//with that account
func GetVcsCreds(repoFullName string, remoteConfig cred.CVRemoteConfig) (*models.VCSCreds, error) {
	vcs := models.NewVCSCreds()
	acctName, _, err := GetAcctRepo(repoFullName)
	if err != nil {
		return nil, err
	}

	bbCreds, err := remoteConfig.GetCredAt(cred.BuildCredPath("bitbucket", acctName, cred.Vcs), false, vcs)
	cf := bbCreds["bitbucket/"+acctName]
	cfg, ok := cf.(*models.VCSCreds)
	// todo: this error happens even if there are no creds there, need a nil check for better error, and also to save to database?? for visibility
	if !ok {
		err = errors.New(fmt.Sprintf("could not cast config as models.VCSCreds, config: %v", cf))
		return nil, err
	}
	return cfg, nil
}

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

	sr := getHookhandlerStageResult(id)
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
func GetBBConfig(remoteConfig cred.CVRemoteConfig, repoFullName string, checkoutCommit string, deserializer *deserialize.Deserializer, vcsHandler handler.VCSHandler) (*pb.BuildConfig, string, error) {
	var bbHandler handler.VCSHandler
	var token string

	if vcsHandler == nil {
		cfg, err := GetVcsCreds(repoFullName, remoteConfig)
		if err != nil {
			ocelog.IncludeErrField(err)
			return nil, "", err
		}

		bbHandler, token, err = handler.GetBitbucketClient(cfg)
		if err != nil {
			ocelog.IncludeErrField(err)
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
		VcsType:      "bitbucket",
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

func getHookhandlerStageResult(id int64) *smods.StageResult {
	start := time.Now()
	return &smods.StageResult{
		Messages: 	   []string{},
		BuildId:       id,
		Stage:         smods.HOOKHANDLER_VALIDATION,
		StartTime:     start,
		StageDuration: -99.99,
	}
}

func BadFormat(msg string) *FormatError {
	return &FormatError{err:msg}
}

type FormatError struct {
	err string
}

func (f *FormatError) Error() string {
	return f.err
}
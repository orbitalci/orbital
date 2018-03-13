package hookhandler

import (
	"bitbucket.org/level11consulting/go-til/deserialize"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	ocevault "bitbucket.org/level11consulting/go-til/vault"
	"bitbucket.org/level11consulting/ocelot/client/validate"
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/handler"
	smods "bitbucket.org/level11consulting/ocelot/util/storage/models"
	"errors"
	"fmt"
	"strings"
	"time"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/util/storage"
)


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

//GetBBConfig returns the protobuf ocelot.yaml, a valid bitbucket token belonging to that repo, and possible err
func GetBBConfig(remoteConfig cred.CVRemoteConfig, repoFullName string, checkoutCommit string, deserializer *deserialize.Deserializer) (conf *pb.BuildConfig, token string, err error) {
	vcs := models.NewVCSCreds()
	acctName, _ := getAcctRepo(repoFullName)

	bbCreds, err := remoteConfig.GetCredAt(cred.BuildCredPath("bitbucket", acctName, cred.Vcs), false, vcs)
	cf := bbCreds["bitbucket/"+acctName]
	cfg, ok := cf.(*models.VCSCreds)
	// todo: this error happens even if there are no creds there, need a nil check for better error, and also to save to database?? for visibility
	if !ok {
		err = errors.New(fmt.Sprintf("could not cast config as models.VCSCreds, config: %v", cf))
		return
	}

	bb, token, err := handler.GetBitbucketClient(cfg)
	if err != nil {
		ocelog.IncludeErrField(err)
		return
	}

	fileBytz, err := bb.GetFile("ocelot.yml", repoFullName, checkoutCommit)
	if err != nil {
		ocelog.IncludeErrField(err)
	}

	conf, err = CheckForBuildFile(fileBytz, deserializer)
	return
}

//QueueAndStore will create a werker task and put it on the queue, then update database
func QueueAndStore(hash, branch, accountRepo, bbToken string,
					messages []string,
					remoteConfig cred.CVRemoteConfig,
					buildConf *pb.BuildConfig,
					validator *validate.OcelotValidator,
					producer *nsqpb.PbProduce,
					store storage.OcelotStorage) error {
	ocelog.Log().Debug("Storing initial results in db")
	account, repo := getAcctRepo(accountRepo)
	vaulty := remoteConfig.GetVault()

	id := storeSummaryToDb(store, hash, repo, branch, account)
	sr := getHookhandlerStageResult(id, messages)
	// stageResult.BuildId, stageResult.Stage, stageResult.Error, stageResult.StartTime, stageResult.StageDuration, stageResult.Status, stageResult.Messages

	if err := validateBuild(buildConf, branch, validator); err == nil {
		tellWerker(buildConf, vaulty, producer, hash, account + "/" + repo, bbToken, id)
		ocelog.Log().Debug("told werker!")
		sr.Status = 0
		sr.Messages = append(sr.Messages, "Passed initial validation " + smods.CHECKMARK)
		//sr.Messages = []string{"Passed initial validation " + smods.CHECKMARK}
	} else {
		sr.Status = 1
		sr.Error = err.Error()
		sr.Messages = append(sr.Messages, "Failed initial validation. Error: " + err.Error())
		//sr.Messages = []string{"Failed initial validation. Error: " + err.Error()}
	}

	sr.StageDuration = time.Now().Sub(sr.StartTime).Seconds()
	if err := storeStageToDb(store, sr); err != nil {
		ocelog.IncludeErrField(err).Error("unable to add hookhandler stage details")
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
	dbid int64) {
	// get one-time token use for access to vault
	token, err := vaulty.CreateThrowawayToken()
	if err != nil {
		ocelog.IncludeErrField(err).Error("unable to create one-time vault token")
		return
	}

	werkerTask := &pb.WerkerTask{
		VaultToken:   token,
		CheckoutHash: hash,
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
func validateBuild(buildConf *pb.BuildConfig, branch string, validator *validate.OcelotValidator) error {
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

// helper
func getAcctRepo(fullName string) (acct string, repo string) {
	list := strings.Split(fullName, "/")
	acct = list[0]
	repo = list[1]
	return
}

func getHookhandlerStageResult(id int64, messages []string) *smods.StageResult {
	start := time.Now()

	return &smods.StageResult{
		Messages: 	   messages,
		BuildId:       id,
		Stage:         smods.HOOKHANDLER_VALIDATION,
		StartTime:     start,
		StageDuration: -99.99,
	}
}

func storeStageToDb(store storage.BuildStage, stageResult *smods.StageResult) error {
	if err := store.AddStageDetail(stageResult); err != nil {
		ocelog.IncludeErrField(err).Error("unable to store hookhandler stage details to db")
		return err
	}
	return nil
}

func storeSummaryToDb(store storage.BuildSum, hash, repo, branch, account string) int64 {
	starttime := time.Now()
	id, err := store.AddSumStart(hash, starttime, account, repo, branch)
	if err != nil {
		ocelog.IncludeErrField(err).Error("unable to store summary details to db")
	}
	return id
}

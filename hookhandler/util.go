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
					remoteConfig cred.CVRemoteConfig,
					buildConf *pb.BuildConfig,
					validator *validate.OcelotValidator,
					producer *nsqpb.PbProduce,
					store storage.OcelotStorage) error {
	ocelog.Log().Debug("Storing initial results in db")
	startTime := time.Now()

	account, repo := getAcctRepo(accountRepo)
	vaulty := remoteConfig.GetVault()

	id := notifyStorage(store, hash, startTime, repo, branch, account)
	sr := getHookhandlerStageResult(startTime, id)
	// stageResult.BuildId, stageResult.Stage, stageResult.Error, stageResult.StartTime, stageResult.StageDuration, stageResult.Status, stageResult.Messages

	if err := ValidateBuild(buildConf, branch, validator); err == nil {
		TellWerker(buildConf, vaulty, producer, hash, account + "/" + repo, bbToken, id)
		ocelog.Log().Debug("told werker!")
		sr.Status = 0
		sr.Messages = []string{"Passed initial validation " + smods.CHECKMARK}
	} else {
		sr.Status = 1
		sr.Error = err.Error()
		sr.Messages = []string{"Failed initial validation. Error: " + err.Error()}
	}

	sr.StageDuration = time.Now().Sub(sr.StartTime).Seconds()
	if err := store.AddStageDetail(sr); err != nil {
		ocelog.IncludeErrField(err).Error("unable to add hookhandler stage details")
		return err
	}
	return nil
}

func TellWerker(buildConf *pb.BuildConfig,
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
func ValidateBuild(buildConf *pb.BuildConfig, branch string, validator *validate.OcelotValidator) error {
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

func getHookhandlerStageResult(start time.Time, id int64) *smods.StageResult {
	return &smods.StageResult{
		BuildId:       id,
		Stage:         smods.HOOKHANDLER_VALIDATION,
		StartTime:     start,
		StageDuration: -99.99, // is this the best way? prob not
	}
}

func notifyStorage(store storage.BuildSum, hash string, starttime time.Time, repo string, branch string, account string) int64 {
	//AddSumStart(hash string, starttime time.Time, account string, repo string, branch string) (int64, error)
	id, err := store.AddSumStart(hash, starttime, account, repo, branch)
	if err != nil {
		ocelog.IncludeErrField(err).Error("could not kick off summary")
	}
	return id
}

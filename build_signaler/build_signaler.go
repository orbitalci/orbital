package build_signaler

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/shankj3/go-til/deserialize"
	"github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"
	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/storage"
)

const SubscriptionUpdateBranch = "master"

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

func storeSummaryToDb(store storage.BuildSum, hash, repo, branch, account string, by pb.SignaledBy, credId int64) (int64, error) {
	id, err := store.AddSumStart(hash, account, repo, branch, by, credId)
	if err != nil {
		log.IncludeErrField(err).Error("unable to store summary details to db")
		return 0, err
	}
	return id, nil
}

//GetConfig returns the protobuf ocelot.yaml, a valid bitbucket token belonging to that repo, and possible err.
//If a VcsHandler is passed, this method will use the existing handler to retrieve the bb config. In that case,
//***IT WILL NOT RETURN A VALID TOKEN FOR YOU - ONLY BUILD CONFIG***
func GetConfig(repoFullName string, checkoutCommit string, deserializer *deserialize.Deserializer, vcsHandler models.VCSHandler) (*pb.BuildConfig, error) {
	//var bbHandler remote.VCSHandler
	//var token string

	if vcsHandler == nil {
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

// CreateOrUpdateSubscription will build out an ActiveSubscription model and then create or update the the subscription in storage.
func CreateOrUpdateSubscription(conf *pb.BuildConfig, signal *Signaler, subscribingAcctRepo string, subscribingVcsType pb.SubCredType) error {
	for _, subscription := range conf.GetSubscriptions() {
		sub := &pb.ActiveSubscription{
			BranchQueueMap: make(map[string]string),
			SubscribedToAcctRepo: subscription.AcctRepo,
			SubscribedToVcsType: subscription.AcctVcsType,
			SubscribingVcsType: subscribingVcsType,
			SubscribingAcctRepo: subscribingAcctRepo,
			Alias: subscription.Alias,
		}
		for _, branchmap := range subscription.GetBranches() {
			split := strings.Split(branchmap, ":")
			if len(split) != 2 {
				return errors.New("unparseable branch mapping")
			}
			upstreamBranch, subscribingBranch := split[0], split[1]
			sub.BranchQueueMap[upstreamBranch] = subscribingBranch
		}
		id, err := signal.Store.InsertOrUpdateActiveSubscription(sub)
		if err != nil {
			log.IncludeErrField(err).WithField("ActiveSubscription", sub).Error("unable to insert active subscription!")
			return errors.Wrap(err, "unable to insert active subscription")
		}
		log.Log().WithField("ActiveSubscription", sub).Infof("successful insert %d", id)
	}
	return nil
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

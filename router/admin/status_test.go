package admin

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"github.com/shankj3/go-til/consul"
	"github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"
)

func TestGuideOcelotServer_GetStatus(t *testing.T) {
	consl := &statusConsl{}
	rc := &credentials.RemoteConfig{Consul:consl}
	store := &statusStore{}
	gos := &guideOcelotServer{Storage:store, RemoteConfig:rc}
	ctx := context.Background()

	consl.inConsul = true
	// hash path first
	status, err := gos.GetStatus(ctx, &pb.StatusQuery{Hash: "1234"})
	if err != nil {
		t.Error(err)
	}
	if status.Stages[0].Messages[0] != "passed first stage, sweet" {
		t.Errorf("wrong first stage returned, first stage is %#v, should have message of 'passed first stage, sweet", status.Stages[0])
	}
	store.failLatest = true
	_, err = gos.GetStatus(ctx, &pb.StatusQuery{Hash: "1234"})
	if err == nil {
		t.Error("storage failed, should return error")
	}
	store.failLatest = false
	store.failStageDet = true
	_, err = gos.GetStatus(ctx, &pb.StatusQuery{Hash: "1234"})
	if err == nil {
		t.Error("storage failed at stage detail retrieve, should return error")
	}
	if !strings.Contains(err.Error(), "failing stage detail") {
		t.Error("wrong error, expected to contain failing stage detail, instead error is: " + err.Error())
	}
	store.failStageDet = false
	consl.returnErr = true
	_, err = gos.GetStatus(ctx, &pb.StatusQuery{Hash: "1234"})
	if err == nil {
		t.Error("consul failed, should return error")
	}
	if !strings.Contains(err.Error(), "An error occurred checking build status in consul") {
		t.Error("wrong error, expected to contain error checking build status, instead error is: " + err.Error())
	}

	// now check by acct name and repo
	query := &pb.StatusQuery{AcctName: "shankj3", RepoName: "ocelot"}
	consl.returnErr = false
	status, err = gos.GetStatus(ctx, query)
	if err != nil {
		t.Error(err)
	}
	if !status.BuildSum.Failed {
		t.Error("processed storage build summary incorrectly")
	}
	store.failLastFew = true
	_, err = gos.GetStatus(ctx, query)
	if err == nil {
		t.Error("storage failed, should bubble up")
	}
	if !strings.Contains(err.Error(), "failing last few sums") {
		t.Error("should have returned error of RetrieveLastFewSums, instead returned " + err.Error())
	}
	store.failLastFew = false
	store.returnMany = true
	_, err = gos.GetStatus(ctx, query)
	if err == nil {
		t.Error("storage returned many summaries, fail")
	}
	if !strings.Contains(err.Error(), "there is no ONE entry that matches the acctname") {
		t.Error("shouldreturn error that there are multiple summaries, got " + err.Error())
	}
	store.returnMany = false
	store.returnNoSums = true
	_, err = gos.GetStatus(ctx, query)
	if err == nil {
		t.Error("storage returned no summaries, should fail")
	}
	if !strings.Contains(err.Error(), "There are no entries that match the acctnam") {
		t.Error("shouldreturn error that there are no summaries, got " + err.Error())
	}
	// partial repo now
	 query = &pb.StatusQuery{PartialRepo: "ocel"}
	 store.returnNoSums = false
	 status, err = gos.GetStatus(ctx, query)
	 if err != nil {
	 	t.Error(err)
	 }
	 if status.Stages[2].StageDuration != 21.17 {
	 	t.Error("got stages out of order")
	 }
	 store.returnMany = true
	 _, err = gos.GetStatus(ctx, query)
	 if err == nil {
	 	t.Error("returned many summaries, should fail")
	 }
	 if !strings.Contains(err.Error(), "there are 2 repositories ") {
	 	t.Error("should return many repos error, returned " + err.Error())
	 }
	 store.returnMany = false
	 store.returnNoSums = true
	_, err = gos.GetStatus(ctx, query)
	if err == nil {
		t.Error("returned no summaries, should fail")
	}
	if !strings.Contains(err.Error(), "there are no repositories starting with ") {
		t.Error("should return no repos error, returned " + err.Error())
	}
	store.returnNoSums = false
	store.failAcctRepo = true
	_, err = gos.GetStatus(ctx, query)
	if err == nil {
		t.Error("failed acct repo retrieval, should fail")
	}
	store.failAcctRepo = false
	store.failLastFew = true
	_, err = gos.GetStatus(ctx, query)
	if err == nil {
		t.Error("failed last few retrieval, should fail")
	}
	store.failLastFew = false
	_, err = gos.GetStatus(ctx, query)
	if err != nil {
		t.Error(err)
	}

	query = &pb.StatusQuery{}
	_, err = gos.GetStatus(ctx, query)
	if err == nil {
		t.Error("no real query sent, should fail")
	}
	if !strings.Contains(err.Error(), "either hash is required, acctName and repoName is required, or partialRepo is required") {
		t.Error("should return validation error, instead returned: " + err.Error())
	}
}

var testSummary = models.BuildSummary{
	Hash: "hashy",
	Failed: true,
	QueueTime: time.Now().Add(-time.Hour),
	BuildTime: time.Now().Add(-time.Hour),
	Account: "shankj3",
	Repo: "ocelot",
	Branch: "master",
	BuildId: 12,
}

var testResults = []models.StageResult{
	{
		BuildId: 12,
		StageResultId: 1,
		Stage: "first",
		Status: int(pb.StageResultVal_PASS),
		Error: "",
		Messages: []string{"passed first stage, sweet"},
		StartTime: time.Now().Add(-time.Minute*30),
		StageDuration: 22.17,
	},
	{
		BuildId: 12,
		StageResultId: 2,
		Stage: "second",
		Status: int(pb.StageResultVal_PASS),
		Error: "",
		Messages: []string{"passed second stage, sweet"},
		StartTime: time.Now().Add(-time.Minute*29),
		StageDuration: 29.17,
	},
	{
		BuildId: 12,
		StageResultId: 3,
		Stage: "third",
		Status: int(pb.StageResultVal_PASS),
		Error: "",
		Messages: []string{"passed third stage, sweet"},
		StartTime: time.Now().Add(-time.Minute*25),
		StageDuration: 21.17,
	},
	{
		BuildId: 12,
		StageResultId: 4,
		Stage: "fourth",
		Status: int(pb.StageResultVal_FAIL),
		Error: "noooo this failed! how dare it!",
		Messages: []string{"failed fourth stage. tsk tsk."},
		StartTime: time.Now().Add(-time.Minute*20),
		StageDuration: 29.17,
	},
}

type statusConsl struct {
	consul.Consuletty
	inConsul bool
	returnErr bool
}

func (s *statusConsl) GetKeyValue(key string) (*api.KVPair, error) {
	if s.returnErr {
		return nil, errors.New("consul error")
	}
	if s.inConsul {
		return &api.KVPair{Key: key, Value: []byte("here i am")}, nil
	}
	return nil, nil
}

type statusStore struct {
	failLatest   bool
	failLastFew  bool
	failAcctRepo bool
	failStageDet bool
	returnNoSums bool
	returnMany   bool
	storage.OcelotStorage
}

func (s *statusStore) RetrieveLastFewSums(repo string, account string, limit int32) ([]models.BuildSummary, error) {
	if s.returnNoSums {
		return []models.BuildSummary{}, nil
	}
	if s.returnMany {
		return []models.BuildSummary{testSummary, testSummary}, nil
	}
	if s.failLastFew {
		return nil, errors.New("failing last few sums")
	}
	return []models.BuildSummary{testSummary}, nil
}

func (s *statusStore) RetrieveAcctRepo(partialRepo string) ([]models.BuildSummary, error) {
	if s.failAcctRepo {
		return nil, errors.New("failing acct repo")
	}
	if s.returnNoSums {
		return []models.BuildSummary{}, nil
	}
	if s.returnMany {
		return []models.BuildSummary{testSummary, testSummary}, nil
	}
	return []models.BuildSummary{testSummary}, nil
}

func (s *statusStore) RetrieveLatestSum(gitHash string) (models.BuildSummary, error) {
	if s.failLatest {
		return models.BuildSummary{}, errors.New("failing latest")
	}
	return testSummary, nil
}

func (s *statusStore) RetrieveStageDetail(buildId int64) ([]models.StageResult, error) {
	if s.failStageDet {
		return nil, errors.New("failing stage detail")
	}
	return testResults, nil
}
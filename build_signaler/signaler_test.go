package build_signaler

import (
	"strings"
	"testing"
	//"time"

	"github.com/go-test/deep"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hashicorp/consul/api"
	"github.com/shankj3/go-til/consul"
	"github.com/shankj3/go-til/deserialize"
	"github.com/shankj3/go-til/test"
	"github.com/shankj3/go-til/vault"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"
)

var goodConfig = &pb.BuildConfig{
	MachineTag: "hi",
	Branches: []string{"branch1", "branch2_.*"},
	BuildTool: "ios",
	Stages: []*pb.Stage{{Name:"test", Script:[]string{"echo hi"}}},
}

var badConfig = &pb.BuildConfig{
	Branches: []string{"branch1", "branch2_.*"},
	MachineTag: "hi",
	Stages: []*pb.Stage{{Name:"test", Script:[]string{"echo hi"}}},
}



func TestSignaler_validateAndQueue(t *testing.T) {
	sig := getFakeSignaler(t, false)
	stageRes := &models.StageResult{}
	err := sig.validateAndQueue(goodConfig, stageRes, "mine", "1234", "token", "jessi/shank", 12)
	if err != nil {
		t.Error(err)
	}
	<-sig.Producer.(*testSingleProducer).done
	expectedBuildMsg := &pb.WerkerTask{VcsType: pb.SubCredType_BITBUCKET, VaultToken: "token", CheckoutHash:"1234", Branch:"mine", BuildConf: goodConfig, VcsToken:"token", FullName:"jessi/shank", Id: 12}
	if diff := deep.Equal(expectedBuildMsg, sig.Producer.(*testSingleProducer).message); diff != nil {
		t.Error(diff)
	}
	if stageRes.Messages[0] != "Passed initial validation "+models.CHECKMARK {
		t.Error(test.StrFormatErrors("stage msg", "Passed initial validation "+models.CHECKMARK, stageRes.Messages[0]))
	}
	// reset
	sig = getFakeSignaler(t, false)
	stageRes = &models.StageResult{}
	err = sig.validateAndQueue(badConfig, stageRes, "mine", "1234", "token", "jessi/shank", 12)
	if err == nil {
		t.Error("invalid build config, should have errored on validate and not queued")
	}
	if stageRes.Messages[0] != "Failed initial validation" {
		t.Error(test.StrFormatErrors("stage result message", "Failed initial validation", strings.Join(stageRes.Messages, ",")))
	}
}

func TestSignaler_queueAndStore(t *testing.T) {
	sig := getFakeSignaler(t, true)
	err := sig.QueueAndStore("1234", "token", "dev", "jessi/shank", goodConfig)
	if err == nil {
		t.Error("build is already in consul, should return an error")
	}
	if _, ok := err.(*build.NotViable); !ok {
		t.Error("build already being in consul should result in a NotViable error")
	}
}

func TestSignaler_queueAndStore_happypath(t *testing.T) {
	sig := getFakeSignaler(t, false)
	err := sig.QueueAndStore("1234", "token", "dev", "jessi/shank", goodConfig)
	if err != nil {
		t.Error("should pass validation and store properly")
	}
	<-sig.Producer.(*testSingleProducer).done
	if sig.Producer.(*testSingleProducer).message == nil || sig.Producer.(*testSingleProducer).topic == "" {
		t.Error("should have produced a message")
	}
	expectedSummary := &pb.BuildSummary{BuildId: 12, Hash: "1234", Branch:"dev", Account:"jessi", Repo:"shank", Failed:false, QueueTime:&timestamp.Timestamp{Seconds:0, Nanos: 0}}
	if diff := deep.Equal(expectedSummary, sig.Store.(*testStorage).summary); diff != nil {
		t.Error(diff)
	}
	expectedStage := &models.StageResult{BuildId:12, Stage: models.HOOKHANDLER_VALIDATION, StageDuration: -99.99, Messages: []string{"Passed initial validation "+models.CHECKMARK}, Status: 0}
	liveStage := sig.Store.(*testStorage).stages[0]
	if expectedStage.BuildId != liveStage.BuildId {
		t.Error(test.GenericStrFormatErrors("build id", expectedStage.BuildId, liveStage.BuildId))
	}
	if expectedStage.Status != liveStage.Status {
		t.Error("stage should be marked as passed")
	}
	if diff := deep.Equal(expectedStage.Messages, liveStage.Messages); diff != nil {
		t.Error(diff)
	}
}

func TestSignaler_queueAndStore_invalid(t *testing.T) {
	sig := getFakeSignaler(t, false)
	err := sig.QueueAndStore("1234", "token", "dev", "jessi/shank", badConfig)
	if err != nil {
		t.Error("should not pass validation, but should not return an error")
	}
	expectedSummary := &pb.BuildSummary{BuildId: 12, Hash: "1234", Branch:"dev", Account:"jessi", Repo:"shank", Failed:true,}
	if diff := deep.Equal(expectedSummary, sig.Store.(*testStorage).summary); diff != nil {
		t.Error(diff)
	}
	expectedStage := &models.StageResult{BuildId:12, Stage: models.HOOKHANDLER_VALIDATION, StageDuration: -99.99, Messages: []string{"Failed initial validation"}, Status: 1}
	liveStage := sig.Store.(*testStorage).stages[0]
	if expectedStage.BuildId != liveStage.BuildId {
		t.Error(test.GenericStrFormatErrors("build id", expectedStage.BuildId, liveStage.BuildId))
	}
	if expectedStage.Status != liveStage.Status {
		t.Error("stage should be marked as failed")
	}
	if diff := deep.Equal(expectedStage.Messages, liveStage.Messages); diff != nil {
		t.Error(diff)
	}
}

func TestSignaler_CheckViableThenQueueAndStore(t *testing.T) {
	sig := getFakeSignaler(t, false)
	err := sig.CheckViableThenQueueAndStore("1234", "token", "slurp", "jessi/shank", badConfig, []*pb.Commit{}, false)
	if err == nil {
		t.Error("branches didn't match list of acceptable branches, this should have failed")
	}
	if _, ok := err.(*build.NotViable); !ok {
		t.Error("branches not matching acceptable branches shuld result in a NotViable error")
	}
}

func getFakeSignaler(t *testing.T, inConsul bool) *Signaler {
	cred := &credentials.RemoteConfig{Consul:&testConsul{keyFound:inConsul}, Vault: &testVault{}}
	dese := deserialize.New()
	valid := &build.OcelotValidator{}
	store := &testStorage{}
	produ := &testSingleProducer{done: make(chan int, 1)}
	return NewSignaler(cred, dese, produ, valid, store)
}


type testSingleProducer struct {
	message proto.Message
	topic string
	done chan int
}

func (tp *testSingleProducer) WriteProto(message proto.Message, topicName string) error {
	tp.message = message
	tp.topic = topicName
	close(tp.done)
	return nil
}

type testVault struct {
	vault.Vaulty
}

func (tv *testVault) CreateThrowawayToken() (string, error) {
	return "token", nil
}

type testConsul struct {
	consul.Consuletty
	keyFound bool
}

func (tc *testConsul) GetKeyValue(string) (*api.KVPair, error) {
	if tc.keyFound {
		return &api.KVPair{}, nil
	}
	return nil, nil
}

// can take one build
type testStorage struct {
	storage.OcelotStorage
	summary *pb.BuildSummary
	stages []*models.StageResult
}

func (ts *testStorage) AddSumStart(hash, account, repo, branch string) (int64, error) {
	ts.summary = &pb.BuildSummary{Hash:hash, Account: account, Repo:repo, Branch:branch, BuildId: 12}
	return 12, nil
}

func (ts *testStorage) SetQueueTime(id int64) error {
	ts.summary.QueueTime = &timestamp.Timestamp{Seconds:0,Nanos:0}
	return nil
}

func (ts *testStorage) StoreFailedValidation(id int64) error {
	ts.summary.Failed = true
	return nil
}

func (ts *testStorage) AddStageDetail(result *models.StageResult) error {
	ts.stages = append(ts.stages, result)
	return nil
}



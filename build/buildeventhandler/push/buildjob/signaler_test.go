package buildjob

import (
	"strings"
	"testing"

	//"time"

	"github.com/go-test/deep"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/level11consulting/ocelot/build"
	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/shankj3/go-til/test"
)

var goodConfig = &pb.BuildConfig{
	MachineTag: "hi",
	Branches:   []string{"branch1", "branch2_.*"},
	BuildTool:  "ios",
	Stages:     []*pb.Stage{{Name: "test", Script: []string{"echo hi"}}},
}

var badConfig = &pb.BuildConfig{
	Branches:   []string{"branch1", "branch2_.*"},
	MachineTag: "hi",
	Stages:     []*pb.Stage{{Name: "test", Script: []string{"echo hi"}}},
}

func TestSignaler_validateAndQueue(t *testing.T) {
	sig := GetFakeSignaler(t, false)
	stageRes := &models.StageResult{}
	task := &pb.WerkerTask{VcsType: pb.SubCredType_BITBUCKET, BuildConf: goodConfig, CheckoutHash: "1234", Branch: "mine", VcsToken: "token", FullName: "jessi/shank", Id: 12}
	err := sig.validateAndQueue(task, stageRes)
	if err != nil {
		t.Error(err)
	}
	<-sig.Producer.(*TestSingleProducer).Done
	expectedBuildMsg := &pb.WerkerTask{VcsType: pb.SubCredType_BITBUCKET, CheckoutHash: "1234", Branch: "mine", BuildConf: goodConfig, VcsToken: "token", FullName: "jessi/shank", Id: 12}
	if diff := deep.Equal(expectedBuildMsg, sig.Producer.(*TestSingleProducer).Message); diff != nil {
		t.Error(diff)
	}
	if stageRes.Messages[0] != "Passed initial validation "+models.CHECKMARK {
		t.Error(test.StrFormatErrors("stage msg", "Passed initial validation "+models.CHECKMARK, stageRes.Messages[0]))
	}
	// reset
	sig = GetFakeSignaler(t, false)
	stageRes = &models.StageResult{}
	task = &pb.WerkerTask{BuildConf: badConfig, CheckoutHash: "1234", Branch: "mine", VcsToken: "token", FullName: "jessi/shank", Id: 12}
	err = sig.validateAndQueue(task, stageRes)
	if err == nil {
		t.Error("invalid build config, should have errored on validate and not queued")
	}
	if stageRes.Messages[0] != "Failed initial validation" {
		t.Error(test.StrFormatErrors("stage result message", "Failed initial validation", strings.Join(stageRes.Messages, ",")))
	}
}

func TestSignaler_queueAndStore(t *testing.T) {
	sig := GetFakeSignaler(t, true)
	task := &pb.WerkerTask{BuildConf: goodConfig, CheckoutHash: "1234", Branch: "dev", VcsToken: "token", FullName: "jessi/shank", Id: 12}
	err := sig.QueueAndStore(task)
	if err == nil {
		t.Error("build is already in consul, should return an error")
	}
	if _, ok := err.(*build.NotViable); !ok {
		t.Error("build already being in consul should result in a NotViable error")
	}
}

func TestSignaler_queueAndStore_happypath(t *testing.T) {
	sig := GetFakeSignaler(t, false)
	task := &pb.WerkerTask{BuildConf: goodConfig, CheckoutHash: "1234", Branch: "dev", VcsToken: "token", FullName: "jessi/shank", Id: 12}
	err := sig.QueueAndStore(task)
	if err != nil {
		t.Error("should pass validation and store properly")
	}
	<-sig.Producer.(*TestSingleProducer).Done
	if sig.Producer.(*TestSingleProducer).Message == nil || sig.Producer.(*TestSingleProducer).Topic == "" {
		t.Error("should have produced a message")
	}
	expectedSummary := &pb.BuildSummary{BuildId: 12, Hash: "1234", Branch: "dev", Account: "jessi", Repo: "shank", Failed: false, QueueTime: &timestamp.Timestamp{Seconds: 0, Nanos: 0}}
	if diff := deep.Equal(expectedSummary, sig.Store.(*TestStorage).summary); diff != nil {
		t.Error(diff)
	}
	expectedStage := &models.StageResult{BuildId: 12, Stage: models.HOOKHANDLER_VALIDATION, StageDuration: -99.99, Messages: []string{"Passed initial validation " + models.CHECKMARK}, Status: 0}
	liveStage := sig.Store.(*TestStorage).stages[0]
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
	sig := GetFakeSignaler(t, false)
	task := &pb.WerkerTask{BuildConf: badConfig, CheckoutHash: "1234", Branch: "dev", VcsToken: "token", FullName: "jessi/shank", Id: 12}
	err := sig.QueueAndStore(task)
	if err != nil {
		t.Error("should not pass validation, but should not return an error")
	}
	expectedSummary := &pb.BuildSummary{BuildId: 12, Hash: "1234", Branch: "dev", Account: "jessi", Repo: "shank", Failed: true}
	if diff := deep.Equal(expectedSummary, sig.Store.(*TestStorage).summary); diff != nil {
		t.Error(diff)
	}
	expectedStage := &models.StageResult{BuildId: 12, Stage: models.HOOKHANDLER_VALIDATION, StageDuration: -99.99, Messages: []string{"Failed initial validation"}, Status: 1}
	liveStage := sig.Store.(*TestStorage).stages[0]
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
	sig := GetFakeSignaler(t, false)
	task := &pb.WerkerTask{BuildConf: badConfig, CheckoutHash: "1234", Branch: "slurp", VcsToken: "token", FullName: "jessi/shank", Id: 12}
	err := sig.CheckViableThenQueueAndStore(task, false, []*pb.Commit{})
	if err == nil {
		t.Error("branches didn't match list of acceptable branches, this should have failed")
	}
	if _, ok := err.(*build.NotViable); !ok {
		t.Error("branches not matching acceptable branches shuld result in a NotViable error")
	}
}

package build_signaler

import (
	"strings"
	"testing"
	//"time"

	"github.com/go-test/deep"
	"github.com/golang/mock/gomock"
	"github.com/shankj3/go-til/consul"
	"github.com/shankj3/go-til/deserialize"
	"github.com/shankj3/go-til/nsqpb"
	"github.com/shankj3/go-til/test"
	"github.com/level11consulting/ocelot/build"
	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/go-til/vault"
	"github.com/level11consulting/ocelot/common/credentials"
	"github.com/level11consulting/ocelot/storage"
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
	ctl := gomock.NewController(t)
	rc := credentials.NewMockCVRemoteConfig(ctl)
	store := storage.NewMockOcelotStorage(ctl)
	producer := nsqpb.NewMockProducer(ctl)
	consu := consul.NewMockConsuletty(ctl)
	sig := &Signaler{
		Producer: producer,
		RC: rc,
		Store: store,
		OcyValidator: build.GetOcelotValidator(),
		Deserializer: deserialize.New(),
	}
	//sig := GetFakeSignaler(t, false)
	// the check and store part
	rc.EXPECT().GetConsul().Return(consu).Times(1)
	// set consul to return that a build is not currently happening
	consu.EXPECT().GetKeyValue(gomock.Any()).Return(nil, nil)
	rc.EXPECT().GetCred(store, pb.SubCredType_BITBUCKET, "BITBUCKET_jessi", "jessi", false).Return(&pb.VCSCreds{Id: int64(1)}, nil).Times(1)
	store.EXPECT().AddSumStart("1234", "jessi", "shank", "dev", pb.SignaledBy_PUSH, int64(1)).Return(int64(2), nil).Times(1)
	task := &pb.WerkerTask{BuildConf: goodConfig, CheckoutHash: "1234", Branch: "dev", FullName: "jessi/shank", VcsType: pb.SubCredType_BITBUCKET, SignaledBy: pb.SignaledBy_PUSH}
	// now the queuing part
	vlt := vault.NewMockVaulty(ctl)
	rc.EXPECT().GetVault().Return(vlt).Times(1)
	vlt.EXPECT().CreateThrowawayToken().Return("token", nil).Times(1)
	producer.EXPECT().WriteProto(&pb.WerkerTask{BuildConf: goodConfig, CheckoutHash: "1234", Branch: "dev", VaultToken: "token", FullName: "jessi/shank", VcsType: pb.SubCredType_BITBUCKET, SignaledBy: pb.SignaledBy_PUSH, Id: 2}, "build_hi").Return(nil).Times(1)
	store.EXPECT().SetQueueTime(int64(2)).Return(nil).Times(1)
	// setting stage detail expect to gomock.Any() because can't predict start time and can't do dthis: &models.StageResult{BuildId: 2, Messages: []string{"Passed initial validation "+models.CHECKMARK}, Status: 0, Stage: "pre-build-validation", StageDuration: gomock.Any()}
	store.EXPECT().AddStageDetail(gomock.Any()).Return(nil).Times(1)
	err := sig.QueueAndStore(task)
	if err != nil {
		t.Error("should pass validation and store properly")
	}
}

func TestSignaler_queueAndStore_invalid(t *testing.T) {
	ctl := gomock.NewController(t)
	rc := credentials.NewMockCVRemoteConfig(ctl)
	store := storage.NewMockOcelotStorage(ctl)
	producer := nsqpb.NewMockProducer(ctl)
	consu := consul.NewMockConsuletty(ctl)
	sig := &Signaler{
		Producer: producer,
		RC: rc,
		Store: store,
		OcyValidator: build.GetOcelotValidator(),
		Deserializer: deserialize.New(),
	}
	//sig := GetFakeSignaler(t, false)
	// the check and store part
	rc.EXPECT().GetConsul().Return(consu).Times(1)
	// set consul to return that a build is not currently happening
	consu.EXPECT().GetKeyValue(gomock.Any()).Return(nil, nil)
	rc.EXPECT().GetCred(store, pb.SubCredType_BITBUCKET, "BITBUCKET_jessi", "jessi", false).Return(&pb.VCSCreds{Id: int64(1)}, nil).Times(1)
	store.EXPECT().AddSumStart("1234", "jessi", "shank", "dev", pb.SignaledBy_PUSH, int64(1)).Return(int64(2), nil).Times(1)
	store.EXPECT().StoreFailedValidation(int64(2)).Return(nil).Times(1)
	// setting stage detail expect to gomock.Any() because can't predict start time and can't do dthis: &models.StageResult{BuildId: 2, Messages: []string{"Passed initial validation "+models.CHECKMARK}, Status: 0, Stage: "pre-build-validation", StageDuration: gomock.Any()}
	store.EXPECT().AddStageDetail(gomock.Any()).Return(nil).Times(1)
	task := &pb.WerkerTask{BuildConf: badConfig, CheckoutHash: "1234", Branch: "dev", VcsToken: "token", FullName: "jessi/shank", VcsType: pb.SubCredType_BITBUCKET, SignaledBy: pb.SignaledBy_PUSH}
	err := sig.QueueAndStore(task)
	if err != nil {
		t.Error("should not pass validation, but should not return an error")
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

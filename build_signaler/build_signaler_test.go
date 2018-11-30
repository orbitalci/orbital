package build_signaler

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/golang/mock/gomock"
	"github.com/shankj3/go-til/deserialize"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"
)

var expectedBuildConf = &pb.BuildConfig{
	BuildTool: "go",
	Image:     "golang:1.10.2-alpine3.7",
	Env:       []string{"BUILD_DIR=/go/src/bitbucket.org/level11consulting/"},
	Branches:  []string{"ALL"},
	Stages: []*pb.Stage{
		{
			Name:   "install consul for testing",
			Script: []string{"apk update", "apk add unzip", "cd /go/bin", "wget https://releases.hashicorp.com/consul/1.1.0/consul_1.1.0_linux_amd64.zip", "echo \"unzipping\"", "unzip consul_1.1.0_linux_amd64.zip", "echo \"Done\""},
		},
		{
			Name:   "configure git",
			Script: []string{`git config --global url."git@bitbucket.org:".insteadOf "https://bitbucket.org/"`},
		},
		{
			Name:   "make stoopid dep thing",
			Script: []string{"mkdir -p $BUILD_DIR", "cp -r $WORKSPACE $BUILD_DIR/go-til"},
		},
		{
			Name:   "install dep & ensure dependencies",
			Script: []string{"cd $BUILD_DIR/go-til", "go get -u github.com/golang/dep/...", "dep ensure -v"},
		},
		{
			Name:   "test",
			Script: []string{"cd $BUILD_DIR", "go test ./..."},
		},
	},
}

func TestCheckForBuildFile(t *testing.T) {
	dese := deserialize.New()
	conf, err := CheckForBuildFile(Buildfile, dese)
	if err != nil {
		t.Error("err")
	}
	if diff := deep.Equal(conf, expectedBuildConf); diff != nil {
		t.Error(diff)
	}
}

func TestGetConfig(t *testing.T) {
	handler := &DummyVcsHandler{Fail: false, Filecontents: Buildfile}
	conf, err := GetConfig("jessi/shank", "12345", deserialize.New(), handler)
	if err != nil {
		t.Error("should not return an error, everything should be fine.")
	}
	if diff := deep.Equal(conf, expectedBuildConf); diff != nil {
		t.Error(diff)
	}
	handler = &DummyVcsHandler{Fail: true}
	_, err = GetConfig("jessi/shank", "12345", deserialize.New(), handler)
	if err == nil {
		t.Error("vcs handler returned a failure, that should be bubbled up")
	}
	_, err = GetConfig("jessi/shank", "12345", deserialize.New(), nil)
	if err == nil {
		t.Error("vcs handler is nil, should return an error")
	}
}


func TestCreateOrModifySubscription(t *testing.T) {
	ctl := gomock.NewController(t)
	store := storage.NewMockOcelotStorage(ctl)
	signal := &Signaler{Store: store}
	confNoSubscriptions := &pb.BuildConfig{}
	acctRepo := "shankj3/ocelot"
	vcsType := pb.SubCredType_GITHUB
	store.EXPECT().DeleteAllActiveSubscriptionsForRepo(acctRepo, vcsType).Return(nil).Times(1)
	if err := CreateOrModifySubscription(confNoSubscriptions, signal, acctRepo, vcsType); err != nil {
		t.Fatal(err)
	}
	confWithOne := &pb.BuildConfig{
		Subscriptions: []*pb.Subscriptions{
			{Alias: "sub1", AcctRepo: "shankj3/sub1", AcctVcsType: pb.SubCredType_GITHUB, Branches: []string{"master:master"}},
		},
	}
	expectedSubscription := &pb.ActiveSubscription{
		BranchQueueMap: map[string]string{"master": "master"},
		SubscribedToVcsType: pb.SubCredType_GITHUB,
		SubscribedToAcctRepo: "shankj3/sub1",
		SubscribingVcsType: pb.SubCredType_GITHUB,
		SubscribingAcctRepo: "shankj3/ocelot",
		Alias: "sub1",
	}
	store.EXPECT().InsertOrUpdateActiveSubscription(expectedSubscription).Return(int64(98), nil).Times(1)
	if err := CreateOrModifySubscription(confWithOne, signal, acctRepo, vcsType); err != nil {
		t.Fatal(err)
	}

	confWithMany := &pb.BuildConfig{
		Subscriptions: []*pb.Subscriptions{
			{Alias: "sub1", AcctRepo: "shankj3/sub1", AcctVcsType: pb.SubCredType_GITHUB, Branches: []string{"master:master"}},
			{Alias: "sub2", AcctRepo: "shankj3/sub2", AcctVcsType: pb.SubCredType_BITBUCKET, Branches: []string{"develop:dev"}},
			{Alias: "sub3", AcctRepo: "level11consulting/sub3", AcctVcsType: pb.SubCredType_BITBUCKET, Branches: []string{"master:master"}},
		},
	}
	expectedSubscriptions := []*pb.ActiveSubscription{
		{
			BranchQueueMap: map[string]string{"master": "master"},
			SubscribedToVcsType: pb.SubCredType_GITHUB,
			SubscribedToAcctRepo: "shankj3/sub1",
			SubscribingVcsType: pb.SubCredType_GITHUB,
			SubscribingAcctRepo: "shankj3/ocelot",
			Alias: "sub1",
		},
		{
			BranchQueueMap: map[string]string{"develop": "dev"},
			SubscribedToVcsType: pb.SubCredType_BITBUCKET,
			SubscribedToAcctRepo: "shankj3/sub2",
			SubscribingVcsType: pb.SubCredType_GITHUB,
			SubscribingAcctRepo: "shankj3/ocelot",
			Alias: "sub2",
		},
		{
			BranchQueueMap: map[string]string{"master": "master"},
			SubscribedToVcsType: pb.SubCredType_BITBUCKET,
			SubscribedToAcctRepo: "level11consulting/sub3",
			SubscribingVcsType: pb.SubCredType_GITHUB,
			SubscribingAcctRepo: "shankj3/ocelot",
			Alias: "sub3",
		},
	}
	for ind, sub := range expectedSubscriptions {
		store.EXPECT().InsertOrUpdateActiveSubscription(sub).Return(int64(ind), nil).Times(1)
	}
	if err := CreateOrModifySubscription(confWithMany, signal, acctRepo, vcsType); err != nil {
		t.Fatal(err)
	}
}
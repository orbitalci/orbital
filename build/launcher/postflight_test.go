package launcher

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/shankj3/go-til/net"
	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/shankj3/go-til/nsqpb"
	"github.com/level11consulting/ocelot/build_signaler/taskbuilder"
	"github.com/level11consulting/ocelot/storage"
)

func TestLauncher_postFlight(t *testing.T) {
	lnchr := &launcher{}
	ctx := context.Background()
	ctl := gomock.NewController(t)
	store := storage.NewMockOcelotStorage(ctl)
	task := &pb.WerkerTask{
		SignaledBy: pb.SignaledBy_PULL_REQUEST,
		FullName:   "level11consulting/ocelot",
		PrData: &pb.PrWerkerData{
			PrId: "2",
			Urls: &pb.PrUrls{
				Approve: "http://approve.me",
			},
		},
		CheckoutHash: "1234",
		Id:           1023,
		VcsType:      pb.SubCredType_BITBUCKET,
	}
	store.EXPECT().FindSubscribeesForRepo(task.FullName, task.VcsType).Return(nil, nil).Times(1)
	lnchr.Store = store
	lnchr.handler = &testHandler{failPostPR: true, cli: &testClient{}}
	err := lnchr.postFlight(ctx, task, true)
	if err == nil || err.Error() != "nah" {
		t.Error("expected error to be bubbled up from handler postPRComment")
	}
	lnchr.handler = &testHandler{cli: &testClient{}}
	err = lnchr.postFlight(ctx, task, true)
	if err != nil {
		t.Error("error should be nil, it is: " + err.Error())
	}
	handler := &testHandler{cli: &testClient{}}
	lnchr.handler = handler
	store.EXPECT().FindSubscribeesForRepo(task.FullName, task.VcsType).Return(nil, nil).Times(1)
	err = lnchr.postFlight(ctx, task, false)
	if err != nil {
		t.Error("error should be nil, it is: " + err.Error())
	}
	if handler.cli.posted != 1 {
		t.Errorf("should have posted to approve, instead was posted %d times", handler.cli.posted)
	}
	task.PrData.Urls.Approve = ""
	handler.cli.posted = 0
	err = lnchr.postFlight(ctx, task, false)
	if err.Error() != "approve url is empty!!" {
		t.Error("should have errored ont he check for the approve url")
	}
	handler.cli.failPostUrl = true
	task.PrData.Urls.Approve = "http://approve.me"
	err = lnchr.postFlight(ctx, task, false)
	if err == nil {
		t.Error("should have returned error from client's PostUrl")
	}
}

func TestLauncher_postFlight_subscriptions(t *testing.T) {
	lnchr := &launcher{}
	ctx := context.Background()
	ctl := gomock.NewController(t)
	store := storage.NewMockOcelotStorage(ctl)
	lnchr.Store = store
	producer := nsqpb.NewMockProducer(ctl)
	lnchr.producer = producer
	lnchr.handler = &testHandler{cli: &testClient{}}
	task := &pb.WerkerTask{
		SignaledBy:   pb.SignaledBy_PUSH,
		FullName:     "level11consulting/ocelot",
		CheckoutHash: "1234",
		Id:           1023,
		Branch:       "master",
		VcsType:      pb.SubCredType_BITBUCKET,
	}
	returnedSubs := []*pb.ActiveSubscription{
		{
			SubscribedToAcctRepo: "level11consulting/ocelot",
			SubscribedToVcsType: pb.SubCredType_BITBUCKET,
			SubscribingAcctRepo: "shankj3/go-til",
			SubscribingVcsType: pb.SubCredType_GITHUB,
			Id: 1,
			Alias: "gotil",
			BranchQueueMap: map[string]string{"master":"master", "develop":"develop"},
		},
		{
			SubscribedToAcctRepo: "level11consulting/ocelot",
			SubscribedToVcsType: pb.SubCredType_BITBUCKET,
			SubscribingAcctRepo: "shankj3/ocytest",
			SubscribingVcsType: pb.SubCredType_BITBUCKET,
			Id: 2,
			Alias: "ocytest",
			BranchQueueMap: map[string]string{"master":"master", "develop":"develop"},
		},
	}
	store.EXPECT().FindSubscribeesForRepo(task.FullName, task.VcsType).Return(returnedSubs, nil).Times(1)
	gotil := &pb.TaskBuilderEvent{Subscription: &pb.UpstreamTaskData{BuildId: 1023, Alias: "gotil", ActiveSubscriptionId: 1}, AcctRepo: "shankj3/go-til", VcsType: pb.SubCredType_GITHUB, Branch: "master", By: pb.SignaledBy_SUBSCRIBED}
	ocytest := &pb.TaskBuilderEvent{Subscription: &pb.UpstreamTaskData{BuildId: 1023, Alias: "ocytest", ActiveSubscriptionId: 2}, AcctRepo: "shankj3/ocytest", VcsType: pb.SubCredType_BITBUCKET, Branch: "master", By: pb.SignaledBy_SUBSCRIBED}
	producer.EXPECT().WriteProto(gotil, taskbuilder.TaskBuilderTopic).Return(nil).Times(1)
	producer.EXPECT().WriteProto(ocytest, taskbuilder.TaskBuilderTopic).Return(nil).Times(1)
	if err := lnchr.postFlight(ctx, task, false); err != nil {
		t.Fatal(err)
	}
	// now if the active build is a rando branch with no matchihng in the branch queue map, the producer shouldn't write any new werker task build events
	task.Branch = "nosubscriptionsBranch"
	store.EXPECT().FindSubscribeesForRepo(task.FullName, task.VcsType).Return(returnedSubs, nil).Times(1)
	if err := lnchr.postFlight(ctx, task, false); err != nil {
		t.Fatal(err)
	}

	// even if one of the task msgs fails, try to run the others.
	task.Branch = "master"
	store.EXPECT().FindSubscribeesForRepo(task.FullName, task.VcsType).Return(returnedSubs, nil).Times(1)
	producer.EXPECT().WriteProto(gotil, taskbuilder.TaskBuilderTopic).Return(errors.New("woops this is bad error no good")).Times(1)
	producer.EXPECT().WriteProto(ocytest, taskbuilder.TaskBuilderTopic).Return(nil).Times(1)
	if err := lnchr.postFlight(ctx, task, false); err != nil {
		t.Fatal(err)
	}
}

func TestLauncher_getAndSetHandler(t *testing.T) {
	lnchr := &launcher{}
	ctx := context.Background()
	err := lnchr.getAndSetHandler(ctx, "", pb.SubCredType_BITBUCKET)
	if err != nil {
		t.Error(err)
	}
	lnchr.handler = nil
	err = lnchr.getAndSetHandler(ctx, "", pb.SubCredType_SSHKEY)
	if err == nil {
		t.Error("sshkey is not a supported subcredtype, should error")
		return
	}
	if err.Error() != "unknown vcs type, cannot create handler with token given" {
		t.Error("should return unkown vcs type error, instead got: " + err.Error())
	}
}

type testHandler struct {
	models.VCSHandler
	failPostPR bool
	cli        *testClient
	posted     int
}

func (t *testHandler) PostPRComment(acctRepo string, prId string, hash string, failed bool, buildId int64) error {
	t.posted++
	if t.failPostPR {
		return errors.New("nah")
	}
	return nil
}

func (t *testHandler) GetClient() net.HttpClient {
	return t.cli
}

type testClient struct {
	posted int
	net.HttpClient
	failPostUrl bool
}

func (t *testClient) PostUrl(url string, body string, unmarshalObj proto.Message) error {
	t.posted++
	if t.failPostUrl {
		return errors.New("nah cli")
	}
	return nil
}

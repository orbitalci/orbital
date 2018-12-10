package launcher

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/shankj3/go-til/net"
	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/models/pb"
)

func TestLauncher_postFlight(t *testing.T) {
	lnchr := &launcher{}
	ctx := context.Background()
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

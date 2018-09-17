package slack

import (
	"net/http"
	"testing"

	"github.com/go-test/deep"
	"github.com/shankj3/ocelot/common/remote/slack"
	"github.com/shankj3/ocelot/models/pb"
)

func TestSlacker_staticstuff(t *testing.T) {
	slacker := Create()
	if slacker.client != http.DefaultClient {
		t.Error("http default client should be the client when slacker is created via the Create() function")
	}
	if slacker.SubType() != pb.SubCredType_SLACK {
		t.Error("sub cred type of slacker should be SLACK")
	}
	if slacker.String() != "slack notification" {
		t.Error("string method of slacker should be 'slack notification'")
	}
}

func Test_determineRelevancy(t *testing.T) {
	notifyList := []pb.StageResultVal{pb.StageResultVal_PASS, pb.StageResultVal_FAIL}
	isWorthy := determineRelevancy(notifyList, pb.BuildStatus_PASSED)
	if !isWorthy {
		t.Error("should return that it is worthy")
	}

	notifyList = []pb.StageResultVal{pb.StageResultVal_FAIL}
	isWorthy = determineRelevancy(notifyList, pb.BuildStatus_PASSED)
	if isWorthy {
		t.Error("should return that it is not worthy")
	}

	isWorthy = determineRelevancy(notifyList, pb.BuildStatus_FAILED)
	if !isWorthy {
		t.Error("should return that it is worthy")
	}
}

func TestSlacker_IsRelevant(t *testing.T) {
	slacker := Create()
	noNotify := &pb.BuildConfig{
		Image: "alpine",
	}
	if slacker.IsRelevant(noNotify, pb.BuildStatus_FAILED) {
		t.Error("should not be relevant if notify isn't instantiated")
	}
	nonotify2 := &pb.BuildConfig{
		Notify: &pb.Notifications{},
	}
	if slacker.IsRelevant(nonotify2, pb.BuildStatus_FAILED) {
		t.Error("should not be relevant because (*Notifications).Slack isnt' instantiated")
	}
	notify := &pb.BuildConfig{
		Notify: &pb.Notifications{
			Slack: &pb.Slack{Identifier: "here", On: []pb.StageResultVal{pb.StageResultVal_FAIL}},
		},
	}
	if !slacker.IsRelevant(notify, pb.BuildStatus_FAILED) {
		t.Error("slacker should be relevant because slack is instantiated")
	}

	if slacker.IsRelevant(notify, pb.BuildStatus_PASSED) {
		t.Error("slacker should not be relevant because even though slack is instantiated, the build passed but the config says only notify on FAIL")
	}
}

func TestSlacker_RunIntegration(t *testing.T) {
	cli := slack.MakeFakePoster(200, "")
	slackCreds := []pb.OcyCredder{
		&pb.NotifyCreds{ClientSecret: "http://slack.test", Identifier: "id1"},
		&pb.NotifyCreds{ClientSecret: "http://slack.rest", Identifier: "id2"},
		&pb.NotifyCreds{ClientSecret: "http://slack.fest", Identifier: "id3"},
	}
	fullResult := &pb.Status{BuildSum: &pb.BuildSummary{Hash: "123", Failed: false}, Stages: []*pb.StageStatus{{StageStatus: "stage1", Error: "", Messages: []string{"good"}}}}
	notifications := &pb.Notifications{
		Slack: &pb.Slack{
			Channel:    "@jessi-shank",
			Identifier: "id2",
			On:         []pb.StageResultVal{pb.StageResultVal_FAIL},
		},
	}
	slacker := &Slacker{client: cli}
	err := slacker.RunIntegration(slackCreds, fullResult, notifications)
	if err != nil {
		t.Error(err)
	}
	if diff := deep.Equal(cli.PostedUrls, []string{"http://slack.rest"}); diff != nil {
		t.Error(diff)
	}
	failcli := slack.MakeFakePoster(http.StatusBadRequest, "this fail though")
	slacker.client = failcli
	if err = slacker.RunIntegration(slackCreds, fullResult, notifications); err == nil {
		t.Fatal("should have failed, as http cli returned a 400")
		return
	}
	expectedErr := slack.WebhookRejected(http.StatusBadRequest, "this fail though")
	if err.Error() != expectedErr.Error() {
		t.Error("error msg should be ", expectedErr.Error(), "it is ", err.Error())
	}

}

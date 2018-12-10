package slack

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/shankj3/go-til/test"
	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/models/pb"
	slack "github.com/level11consulting/ocelot/models/slack/pb"
)

func TestThrowStatusWebhook(t *testing.T) {
	url := "https://hooks.slack.com/fake"
	status := &pb.Status{
		BuildSum: &pb.BuildSummary{
			BuildId: 1234,
			Failed:  true,
			Hash:    "testhash",
			Branch:  "banana",
			Account: "jessishank",
			Repo:    "ocyocyocyocy",
			Status:  pb.BuildStatus_FAILED,
		},
		Stages: []*pb.StageStatus{
			{
				Error:       "all good here",
				StageStatus: "prebuild",
				Messages:    []string{"YOU PASSED YOU RAMBUNCTIOUS FELLA"},
				Status:      int32(pb.StageResultVal_PASS),
			},
			{
				Error:       "this has failed!",
				StageStatus: "buildmeeee",
				Messages:    []string{"it failed because you are a failure " + models.FAILED},
				Status:      int32(pb.StageResultVal_FAIL),
			},
		},
	}
	channel := ""
	fakeCli := &FakePoster{ResponseCode: http.StatusOK}
	// uncomment below and set url to a real slack url to see what it looks like
	//url = "<real webhook url>"
	//err := ThrowStatusWebhook(http.DefaultClient, url, channel, status, "https://ocelot.metaverse.l11.com")
	err := ThrowStatusWebhook(fakeCli, url, channel, status, "https://ocelot.me")
	if err != nil {
		t.Error(err)
	}
	expected := &slack.WebhookMsg{
		Username: "ocelot",
		IconUrl:  ocelotIcon,
		Attachments: []*slack.Attachment{
			{
				Fallback: "Build for `jessishank/ocyocyocyocy` at commit `testhash` and branch `banana` has *failed*.\n Build Id is 1234. \nExecute `ocelot logs -build-id 1234` in a terminal for more information.\nYou can also visit https://ocelot.me/repos/jessishank/ocyocyocyocy/1234",
				Color:    "danger",
				Pretext:  "*Ocelot Status*",
				Title:    "Build failed",
				Text:     "Stage details: \n```\n[prebuild] Passed\n\t * YOU PASSED YOU RAMBUNCTIOUS FELLA: all good here\n[buildmeeee] Failed\n\t * it failed because you are a failure " + models.FAILED + ": this has failed!```\n",
				Fields: []*slack.Field{
					{Title: "Repo", Value: "jessishank/ocyocyocyocy", Short: false},
					{Title: "Branch", Value: "banana", Short: true},
					{Title: "Commit", Value: "testhas", Short: true},
					{Title: "Logs Command", Value: "`ocelot logs -build-id 1234`", Short: false},
					{Title: "Detail Url", Value: "https://ocelot.me/repos/jessishank/ocyocyocyocy/1234", Short: false},
				},
			},
		},
	}
	expectedBits, err := json.Marshal(expected)
	if err != nil {
		t.Error(err)
		return
	}
	if !bytes.Equal(expectedBits, fakeCli.PostBody) {
		t.Errorf("webhooks not equal:\nexpected is: \n%s\n\nlive is:\n%s", string(expectedBits), string(fakeCli.PostBody))
	}
}

func TestThrowStatusWebhook_pass(t *testing.T) {
	url := "https://hooks.slack.com/fake"
	fakeCli := &FakePoster{ResponseCode: http.StatusOK}
	status := &pb.Status{
		BuildSum: &pb.BuildSummary{
			BuildId: 1234,
			Failed:  false,
			Hash:    "testhash",
			Branch:  "banana",
			Account: "jessishank",
			Repo:    "ocyocyocyocy",
		},
		Stages: []*pb.StageStatus{
			{
				Error:       "all good here",
				StageStatus: "prebuild",
				Messages:    []string{"YOU PASSED YOU RAMBUNCTIOUS FELLA"},
				Status:      int32(pb.StageResultVal_PASS),
			},
			{
				Error:       "this has failed!",
				StageStatus: "buildmeeee",
				Messages:    []string{"it failed because you are a failure " + models.FAILED},
				Status:      int32(pb.StageResultVal_PASS),
			},
		},
	}
	channel := "@jessi-shank"
	err := ThrowStatusWebhook(fakeCli, url, channel, status, "")
	expected := &slack.WebhookMsg{
		Username: "ocelot",
		IconUrl:  ocelotIcon,
		Channel:  "@jessi-shank",
		Attachments: []*slack.Attachment{
			{
				Fallback: "Build for `jessishank/ocyocyocyocy` at commit `testhash` and branch `banana` has *passed*.\n Build Id is 1234. \nExecute `ocelot logs -build-id 1234` in a terminal for more information.",
				Color:    "good",
				Pretext:  "*Ocelot Status*",
				Title:    "Build passed",
				Text:     "Stage details: \n```\n[prebuild] Passed\n[buildmeeee] Passed```\n",
				Fields: []*slack.Field{
					{Title: "Repo", Value: "jessishank/ocyocyocyocy", Short: false},
					{Title: "Branch", Value: "banana", Short: true},
					{Title: "Commit", Value: "testhas", Short: true},
					{Title: "Logs Command", Value: "`ocelot logs -build-id 1234`", Short: false},
				},
			},
		},
	}
	expectedBits, err := json.Marshal(expected)
	if err != nil {
		t.Error(err)
		return
	}
	if !bytes.Equal(expectedBits, fakeCli.PostBody) {
		t.Errorf("webhooks not equal:\nexpected is: \n%s\n\nlive is:\n%s", string(expectedBits), string(fakeCli.PostBody))
	}
}

func TestThrowStatusWebhook_handleError(t *testing.T) {
	url := "https://hooks.slack.com/fake"
	fakeCli := &FakePoster{ResponseCode: http.StatusBadRequest, ResponseBody: "bad job guy"}
	status := &pb.Status{
		BuildSum: &pb.BuildSummary{
			BuildId: 1234,
			Failed:  false,
			Hash:    "testhash",
			Branch:  "banana",
			Account: "jessishank",
			Repo:    "ocyocyocyocy",
		},
		Stages: []*pb.StageStatus{
			{
				Error:       "all good here",
				StageStatus: "prebuild",
				Messages:    []string{"YOU PASSED YOU RAMBUNCTIOUS FELLA"},
				Status:      int32(pb.StageResultVal_PASS),
			},
			{
				Error:       "this has failed!",
				StageStatus: "buildmeeee",
				Messages:    []string{"it failed because you are a failure " + models.FAILED},
				Status:      int32(pb.StageResultVal_PASS),
			},
		},
	}
	err := ThrowStatusWebhook(fakeCli, url, "", status, "")
	if err == nil {
		t.Error("should be an error, return status of 400")
		return
	}
	if _, ok := err.(*WebhookRejectedErr); !ok {
		t.Error(err)
		return
	}
	if err.Error() != "received a 400, error is: bad job guy" {
		t.Error(test.StrFormatErrors("err msg", "received a 400, error is: bad job guy", err.Error()))
	}
}

package slack

import (
	"net/http"
	"testing"

	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
)

func TestThrowStatusWebhook(t *testing.T) {
	url := "https://hooks.slack.com/services/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"
	status := &pb.Status{
		BuildSum: &pb.BuildSummary{
			Failed: true,
			Hash: "testhash",
			Branch: "banana",
			Account: "jessishank",
			Repo: "ocyocyocyocy",
		},
		Stages: []*pb.StageStatus{
			{
				Error: "this has failed!",
				StageStatus: "buildmeeee",
				Messages: []string{"it failed because you are a failure " + models.FAILED},
				Status: int32(pb.StageResultVal_FAIL),

			},
		},
	}
	channel := ""
	err := ThrowStatusWebhook(http.DefaultClient, url, channel, status)
	if err != nil {
		t.Error(err)
	}
}

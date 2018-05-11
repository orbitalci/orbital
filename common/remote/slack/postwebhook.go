package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/shankj3/ocelot/models/pb"
	slack "github.com/shankj3/ocelot/models/slack/pb"
)

const ocelotIcon = "https://78.media.tumblr.com/avatar_06e2167f3e45_128.pnj"

// ThrowStatusWebhook will create a status string from the protobuf message Status as defined in guideocelot.proto and
//   will post the data to the slack url provided. If the status code is not 200 OK, then a WebhookRejectedErr will be generated and the error
// 	 body will contain the error returned from the slack api.
func ThrowStatusWebhook(cli Poster, url string, channel string, results *pb.Status) error {
	var status string
	var color string
	if results.BuildSum.Failed {
		status = "failed"
		color = "danger"
	} else {
		status = "passed"
		color = "good"
	}
	// header is a fallback for if attachments can't be rendered properly
	header := fmt.Sprintf("Build for `%s/%s` at commit `%s` and branch `%s` has *%s*.\n Build Id is %d. \n", results.BuildSum.Account, results.BuildSum.Repo, results.BuildSum.Hash, results.BuildSum.Branch, status, results.BuildSum.BuildId)
	mid := "Stage details: \n"
	var stageStatus = "```"
	if results != nil && len(results.Stages) > 0 {
		for _, stage := range results.Stages {
			var stageStatusStr string
			if stage.Status == 0 {
				stageStatusStr = "Passed"
			} else {
				stageStatusStr = "Failed"
			}
			stageStatus += fmt.Sprintf("\n[%s] %s", stage.StageStatus, stageStatusStr)
			if results.BuildSum.Failed {
				stageStatus += fmt.Sprintf("\n\t * %s", strings.Join(stage.Messages, "\n\t * "))
				if len(stage.Error) > 0 {
					stageStatus += fmt.Sprintf(": %s", stage.Error)
				}
			}
		}
		stageStatus += "```\n"
	}
	runCommand := fmt.Sprintf("Execute `ocelot logs -build-id %d` in a terminal for more information.", results.BuildSum.BuildId)
	fallback := header + runCommand
	combined := mid + stageStatus + runCommand
	postMsg := &slack.WebhookMsg{
		Username: "ocelot",
		IconUrl: ocelotIcon,
		Attachments: []*slack.Attachment{
			{
				Fallback: fallback,
				Color: color,
				Pretext: "*Ocelot Status*",
				Title: "Build " + status,
				Text: combined,
				Fields: []*slack.Field{
					{"Repo", fmt.Sprintf("%s/%s", results.BuildSum.Account, results.BuildSum.Repo), false},
					{"Branch", results.BuildSum.Branch, true},
					{"Commit", results.BuildSum.Hash[:7], true},
				},
			},
		},
	}
	if channel != "" {
		postMsg.Channel = channel
	}
	postBytes, err := json.Marshal(postMsg)
	if err != nil {
		return err
	}
	resp, err := cli.Post(url, "application/json", bytes.NewReader(postBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// don't bother to read body if everything is good
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	// slack errors: https://api.slack.com/changelog/2016-05-17-changes-to-errors-for-incoming-webhooks
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return WebhookRejected(resp.StatusCode, string(body))
}

// WebhookRejected will return a RejectedError with the reason as the message to be returned by a call to Error()
func WebhookRejected(statusCode int, errorMsg string) *WebhookRejectedErr {
	return &WebhookRejectedErr{msg: fmt.Sprintf("received a %d, error is: %s", statusCode, errorMsg)}
}

type WebhookRejectedErr struct {
	msg string
}

func(r *WebhookRejectedErr) Error() string {
	return r.msg
}
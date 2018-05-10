package slack

import (
	"net/http"

	"github.com/shankj3/ocelot/common/remote/slack"
	"github.com/shankj3/ocelot/models/pb"
)
func Create() *Slacker {
	return &Slacker{client: http.DefaultClient}
}

type Slacker struct {
	client slack.Poster
}

func (s *Slacker) SubType() pb.SubCredType {
	return pb.SubCredType_SLACK
}

func (s *Slacker) IsRelevant(wc *pb.BuildConfig) bool {
	if wc.Notify != nil {
		if wc.Notify.Slack != nil {
			return true
		}
	}
	return false
}

func (s *Slacker) String() string {
	return "slack notification"
}

func (s *Slacker) RunIntegration(slackCreds []pb.OcyCredder, fullResult *pb.Status, notifications *pb.Notifications) error {
	for _, slackCred := range slackCreds {
		if slackCred.GetIdentifier() == notifications.Slack.Identifier {
			err := slack.ThrowStatusWebhook(s.client, slackCred.GetClientSecret(), notifications.Slack.Channel, fullResult)
			if err != nil {
				return err
			}
		}
	}
	return nil
}


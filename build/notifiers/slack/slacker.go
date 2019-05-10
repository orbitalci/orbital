package slack

import (
	"net/http"

	"github.com/level11consulting/orbitalci/models/pb"
)

func Create() *Slacker {
	return &Slacker{client: http.DefaultClient}
}

type Slacker struct {
	client Poster
}

func (s *Slacker) SubType() pb.SubCredType {
	return pb.SubCredType_SLACK
}

func determineRelevancy(notifies []pb.StageResultVal, status pb.BuildStatus) (isWorthy bool) {
	var notifyPass pb.StageResultVal
	if status == pb.BuildStatus_FAILED {
		notifyPass = pb.StageResultVal_FAIL
	} else {
		notifyPass = pb.StageResultVal_PASS
	}
	for _, notifyAcceptable := range notifies {
		if notifyAcceptable == notifyPass {
			return true
		}
	}
	return false
}

func (s *Slacker) IsRelevant(wc *pb.BuildConfig, buildStatus pb.BuildStatus) bool {
	if wc.Notify != nil {
		if wc.Notify.Slack != nil {
			if isWorthy := determineRelevancy(wc.Notify.Slack.On, buildStatus); !isWorthy {
				return false
			}
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
		slacky := slackCred.(*pb.NotifyCreds)
		if slackCred.GetIdentifier() == notifications.Slack.Identifier {
			err := ThrowStatusWebhook(s.client, slackCred.GetClientSecret(), notifications.Slack.Channel, fullResult, slacky.DetailUrlBase)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

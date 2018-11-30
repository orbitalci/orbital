package launcher

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/shankj3/go-til/log"
	"github.com/shankj3/ocelot/common/remote"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
)

// getAndSetHandler will use the accesstoken and vcstype to generate a handler without autorefresh capability and set it to (*launcher).handler field. if (*launcher).handler is already set, will do nothing
func (w *launcher) getAndSetHandler(ctx context.Context, accessToken string, vcsType pb.SubCredType) (err error) {
	if w.handler == nil {
		var handler models.VCSHandler
		handler, err = remote.GetHandlerWithToken(ctx, accessToken, vcsType)
		if err != nil {
			return
		}
		w.handler = handler
	}
	return
}

// postFlight is what the launcher will do at the conclusion of the build, after all stages have run and everything is stored.
//   currently, if the build was signalled by a Pull Request, then a comment will be added to the PR in the subsequent VCS
func (w *launcher) postFlight(ctx context.Context, werk *pb.WerkerTask, failed bool) (err error) {
	if werk.SignaledBy == pb.SignaledBy_PULL_REQUEST {
		err = w.getAndSetHandler(ctx, werk.VcsToken, werk.VcsType)
		if err != nil {
			return
		}
		if err = w.handler.PostPRComment(werk.FullName, werk.PrData.PrId, werk.CheckoutHash, failed, werk.Id); err != nil {
			return
		}
		// only do approve for now, because decline is irreversible i guess??
		if !failed {
			if werk.PrData.Urls.Approve == "" {
				return errors.New("approve url is empty!!")
			}
			err = w.handler.GetClient().PostUrl(werk.PrData.Urls.Approve, "{}", nil)
			if err != nil {
				return
			}
		}
	}

	subscribees, err := w.Store.FindSubscribeesForRepo(werk.FullName, werk.VcsType)
	if err != nil {
		return errors.Wrap(err, "unable to find subscribees for repo")
	}
	for _, subscribee := range subscribees {
		branchToQueue, ok := subscribee.BranchQueueMap[werk.Branch]
		if !ok {
			// the current building branch is not in the list of branches to trigger a downstream build off of, so don't to anything
			continue
		}
		//_ = fmt.Sprintf(branchToQueue)
		log.Log().WithField("activeSubscription", subscribee).Info("found a subscribing account repo to this build/branch")
		taskBuilderData := &pb.TaskBuilderEvent{
			Subscription: &pb.UpstreamTaskData{BuildId: werk.Id, ActiveSubscriptionId: subscribee.Id, Alias: subscribee.Alias},
			AcctRepo: subscribee.SubscribingAcctRepo,
			VcsType: subscribee.SubscribingVcsType,
			Branch: branchToQueue,
			By: pb.SignaledBy_SUBSCRIBED,
		}
		if err = w.producer.WriteProto(taskBuilderData, "taskbuilder"); err != nil {
			log.IncludeErrField(err).WithField("activeSubscription", subscribee).Error("unable to write to task builder queue for building our a werker task")
		}
		_ = fmt.Sprintf("%#v", taskBuilderData)
	}
	return nil
}

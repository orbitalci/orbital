package launcher

import (
	"context"

	"github.com/pkg/errors"
	"github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/common/remote"
	"github.com/shankj3/ocelot/models/pb"
)

// postFlight is what the launcher will do at the conclusion of the build, after all stages have run and everything is stored.
//   currently, if the build was signalled by a Pull Request, then a comment will be added to the PR in the subsequent VCS
func (w *launcher) postFlight(ctx context.Context, werk *pb.WerkerTask, failed bool) error {
	if werk.SignaledBy == pb.SignaledBy_PULL_REQUEST {
		creds, err := credentials.GetVcsCreds(w.Store, werk.FullName, w.RemoteConf)
		if err != nil {
			return err
		}
		handler,_,  err := remote.GetHandler(creds)
		if err != nil {
			return err
		}
		if err = handler.PostPRComment(werk.FullName, werk.PrData.PrId, werk.CheckoutHash, failed, werk.Id); err != nil {
			return err
		}
		// only do approve for now, because decline is irreversible i guess??
		if !failed {
			if werk.PrData.Urls.Approve  == "" {
				return errors.New("approve url is empty!!")
			}
			err = handler.GetClient().PostUrl(werk.PrData.Urls.Approve, "{}", nil)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
package poll

import (
	"errors"

	ocelog "github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"

	"github.com/shankj3/ocelot/build"
	sig "github.com/shankj3/ocelot/build_signaler"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
)

type CCWerkerTeller struct{}

func (w *CCWerkerTeller) TellWerker(hash string, signaler *sig.Signaler, branch string, handler models.VCSHandler, token, acctRepo string, commits []*pb.Commit, force bool) (err error) {
	ocelog.Log().WithField("hash", hash).WithField("acctRepo", acctRepo).WithField("branch", branch).Info("found new commit")
	if token == "" {
		return errors.New("token cannot be empty")
	}
	var buildConf *pb.BuildConfig
	buildConf, err = sig.GetConfig(acctRepo, hash, signaler.Deserializer, handler)
	if err != nil {
		if err == ocenet.FileNotFound {
			return errors.New("no ocelot yaml found for repo " + acctRepo)
		}
		return errors.New("unable to get build configuration; err: " + err.Error())
	}

	if err = signaler.CheckViableThenQueueAndStore(hash, token, branch, acctRepo, buildConf, commits, false); err != nil {
		if _, ok := err.(*build.NotViable); ok {
			return errors.New("did not queue because it shouldn't be queued. explanation: " + err.Error())
		}
		return errors.New("unable to queue or store; err: " + err.Error())
	}
	return nil
}


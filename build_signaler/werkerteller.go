package build_signaler

import (
	"errors"

	ocelog "github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
)


// made this interface for easy testing
type WerkerTeller interface {
	TellWerker(lastCommit string, conf *Signaler, branch string, handler models.VCSHandler, token, acctRepo string) (err error)
}

type VcsWerkerTeller struct{}

func (w *VcsWerkerTeller) TellWerker(hash string, conf *Signaler, branch string, handler models.VCSHandler, token, acctRepo string) (err error) {
	ocelog.Log().WithField("hash", hash).WithField("acctRepo", acctRepo).WithField("branch", branch).Info("found new commit")
	if token == "" {
		return errors.New("token cannot be empty")
	}
	var buildConf *pb.BuildConfig
	buildConf, err = GetConfig(acctRepo, hash, conf.Deserializer, handler)
	if err != nil {
		if err == ocenet.FileNotFound {
			return errors.New("no ocelot yaml found for repo " + acctRepo)
		}
		return errors.New("unable to get build configuration; err: " + err.Error())
	}
	if err = conf.CheckViableThenQueueAndStore(hash, token, branch, acctRepo, buildConf); err != nil {
		if _, ok := err.(*build.NotViable); ok {
			return errors.New("did not queue because it shouldn't be queued. explanation: " + err.Error())
		}
		return errors.New("unable to queue or store; err: " + err.Error())
	}
	return nil
}

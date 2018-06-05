package build_signaler

import (
	"errors"

	"github.com/shankj3/go-til/deserialize"
	ocelog "github.com/shankj3/go-til/log"
	ocenet "github.com/shankj3/go-til/net"
	"github.com/shankj3/go-til/nsqpb"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"
)

type Signaler struct {
	RC credentials.CVRemoteConfig
	*deserialize.Deserializer
	Producer     *nsqpb.PbProduce
	OcyValidator *build.OcelotValidator
	Store        storage.OcelotStorage
	AcctRepo     string
}

// made this interface for easy testing
type WerkerTeller interface {
	TellWerker(lastCommit string, conf *Signaler, branch string, handler models.VCSHandler, token string) (err error)
}

type BBWerkerTeller struct{}

func (w *BBWerkerTeller) TellWerker(hash string, conf *Signaler, branch string, handler models.VCSHandler, token string) (err error) {
	ocelog.Log().WithField("hash", hash).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Info("found new commit")
	if token == "" {
		return errors.New("token cannot be empty")
	}
	var buildConf *pb.BuildConfig
	buildConf, err = GetConfig(conf.AcctRepo, hash, conf.Deserializer, handler)
	if err != nil {
		if err == ocenet.FileNotFound {
			return errors.New("no ocelot yaml found for repo " + conf.AcctRepo)
		}
		return errors.New("unable to get build configuration; err: " + err.Error())
	}
	if err = QueueAndStore(hash, branch, conf.AcctRepo, token, conf.RC, buildConf, conf.OcyValidator, conf.Producer, conf.Store); err != nil {
		if _, ok := err.(*build.DoNotQueue); ok {
			return errors.New("did not queue because it shouldn't be queued. explanation: " + err.Error())
		}
		return errors.New("unable to queue or store; err: " + err.Error())
	}
	return nil
}

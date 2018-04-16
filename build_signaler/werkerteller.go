package build_signaler

import (
	"errors"

	"bitbucket.org/level11consulting/go-til/deserialize"
	ocelog "bitbucket.org/level11consulting/go-til/log"
	ocenet "bitbucket.org/level11consulting/go-til/net"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	"bitbucket.org/level11consulting/ocelot/build"
	"bitbucket.org/level11consulting/ocelot/common/credentials"
	"bitbucket.org/level11consulting/ocelot/common/remote"
	"bitbucket.org/level11consulting/ocelot/models/pb"
	"bitbucket.org/level11consulting/ocelot/storage"
)


type Signaler struct {
	RC credentials.CVRemoteConfig
	*deserialize.Deserializer
	Producer     *nsqpb.PbProduce
	OcyValidator *build.OcelotValidator
	Store        storage.OcelotStorage
	AcctRepo  	  string
}


// made this interface for easy testing
type WerkerTeller interface {
	TellWerker(lastCommit string, conf *Signaler, branch string, handler remote.VCSHandler, token string) (err error)
}

type BBWerkerTeller struct {}

func (w *BBWerkerTeller) TellWerker(hash string, conf *Signaler, branch string, handler remote.VCSHandler, token string) (err error){
	ocelog.Log().WithField("hash", hash).WithField("acctRepo", conf.AcctRepo).WithField("branch", branch).Info("found new commit")
	var buildConf *pb.BuildConfig
	var tokenPulled string
	buildConf, tokenPulled, err = GetBBConfig(conf.RC, conf.Store, conf.AcctRepo, hash, conf.Deserializer, handler)
	if token == "" { token = tokenPulled };
	if err != nil {
		if err == ocenet.FileNotFound {
			return errors.New("no ocelot yaml found for repo " + conf.AcctRepo)
		}
		return errors.New("unable to get build configuration; err: " + err.Error())
	}
	if err = QueueAndStore(hash, branch, conf.AcctRepo, token, conf.RC, buildConf, conf.OcyValidator, conf.Producer, conf.Store); err != nil {
		return errors.New("unable to queue or store; err: " + err.Error())
	}
	return nil
}

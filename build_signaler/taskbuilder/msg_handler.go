package taskbuilder

import (
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/shankj3/go-til/deserialize"
	"github.com/shankj3/go-til/log"
	"github.com/shankj3/go-til/nsqpb"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/common/remote"
	"github.com/shankj3/ocelot/storage"

	signal "github.com/shankj3/ocelot/build_signaler"
	"github.com/shankj3/ocelot/models/pb"
)
func NewMsgHandler(topic string, rc credentials.CVRemoteConfig, store storage.OcelotStorage, producer nsqpb.Producer) *MsgHandler {
	return &MsgHandler{
		Topic:   topic,
		Signaler: signal.NewSignaler(rc, deserialize.New(), producer, build.GetOcelotValidator(), store),
	}
}

type MsgHandler struct {
	Topic   	 string
	*signal.Signaler
}

func (m *MsgHandler) UnmarshalAndProcess(msg []byte, done chan int, finish chan int) error {
	log.Log().Debug("unmarshaling and processing a poll update msg")
	taskBuild := &pb.TaskBuilderEvent{}
	defer func() {
		done <- 1
	}()
	if err := proto.Unmarshal(msg, taskBuild); err != nil {
		log.IncludeErrField(err).Error("unmarshal error for task builder msg")
		return err
	}
	var err error
	switch m.Topic {
	case "taskbuilder":
		vcs, err := credentials.GetVcsCreds(m.Store, taskBuild.AcctRepo, m.RC, taskBuild.VcsType)
		if err != nil {
			log.IncludeErrField(err).WithField("acctRepo", taskBuild.AcctRepo).WithField("vcsType", taskBuild.VcsType.String()).WithField("branch", taskBuild.Branch).Errorf("unable to get vcs creds")
			return err
		}
		handler, token, err := remote.GetHandler(vcs)
		if err != nil {
			log.IncludeErrField(err).WithField("acctRepo", taskBuild.AcctRepo).WithField("vcsType", taskBuild.VcsType.String()).WithField("branch", taskBuild.Branch).Error("unable to get remote vcs handler")
			return err
		}
		hist, err := handler.GetBranchLastCommitData(taskBuild.AcctRepo, taskBuild.Branch)
		if err != nil {
			log.IncludeErrField(err).WithField("acctRepo", taskBuild.AcctRepo).WithField("vcsType", taskBuild.VcsType.String()).WithField("branch", taskBuild.Branch).Error("unable to get last commit data")
			return err
		}
		taskBuild.Hash = hist.Hash
		buildConf, err := signal.GetConfig(taskBuild.AcctRepo, taskBuild.Hash, m.Deserializer, handler)
		if err != nil {
			log.IncludeErrField(err).WithField("acctRepo", taskBuild.AcctRepo).WithField("vcsType", taskBuild.VcsType.String()).WithField("branch", taskBuild.Branch).Error("unable to get build config (ocelot.yml)")
			return err
		}
		task := signal.BuildInitialWerkerTask(buildConf, taskBuild.Hash, token, taskBuild.Branch, taskBuild.AcctRepo, taskBuild.By, nil, handler.GetVcsType())
		//fixme we can't actually use QueueAndStore, because we need to intercept the build id right after it stores and _before_ it queues. we will be storing that build id
		// in the subscription_data table, so that when the build is queued and picked up the werker can query the table to get all the environment variables of the upstream build.
		if err := m.Signaler.QueueAndStore(task); err != nil {
			if _, ok := err.(*build.NotViable); !ok {
				log.IncludeErrField(err).WithField("acctRepo", taskBuild.AcctRepo).WithField("vcsType", taskBuild.VcsType.String()).WithField("branch", taskBuild.Branch).Error("unable to queue and store")
				return err
			}
			log.Log().Info("not queuing because i'm not supposed to, explanation: " + err.Error())
		}
	default:
		err = errors.New("only supported topics are poll_please and no_poll_please")
		log.IncludeErrField(err).Error()
	}
	return err
}

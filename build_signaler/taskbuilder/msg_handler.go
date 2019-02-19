package taskbuilder

import (
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/shankj3/go-til/deserialize"
	"github.com/shankj3/go-til/log"
	"github.com/shankj3/go-til/nsqpb"
	"github.com/level11consulting/ocelot/build"
	"github.com/level11consulting/ocelot/common/credentials"
	"github.com/level11consulting/ocelot/common/remote"
	"github.com/level11consulting/ocelot/storage"

	signal "github.com/level11consulting/ocelot/build_signaler"
	"github.com/level11consulting/ocelot/models/pb"
)

const (
	TaskBuilderTopic = "taskbuilder"
	TaskBuilderChannel = "taskbuilder-channel"
)

func NewMsgHandler(topic string, rc credentials.CVRemoteConfig, store storage.OcelotStorage, producer nsqpb.Producer)  *MsgHandler {
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
		logWithFields := log.Log().WithField("acctRepo", taskBuild.AcctRepo).WithField("vcsType", taskBuild.VcsType.String()).WithField("branch", taskBuild.Branch).WithField("subscription", taskBuild.Subscription)
		vcs, err := credentials.GetVcsCreds(m.Store, taskBuild.AcctRepo, m.RC, taskBuild.VcsType)
		if err != nil {
			logWithFields.WithField("error", err).Errorf("unable to get vcs creds")
			return err
		}
		handler, token, err := remote.GetHandler(vcs)
		if err != nil {
			logWithFields.WithField("error", err).Error("unable to get remote vcs handler")
			return err
		}
		hist, err := handler.GetBranchLastCommitData(taskBuild.AcctRepo, taskBuild.Branch)
		if err != nil {
			logWithFields.WithField("error", err).Error("unable to get last commit data")
			return err
		}
		taskBuild.Hash = hist.Hash
		buildConf, err := signal.GetConfig(taskBuild.AcctRepo, taskBuild.Hash, m.Deserializer, handler)
		if err != nil {
			logWithFields.WithField("error", err).Error("unable to get build config (ocelot.yml)")
			return err
		}
		task := signal.BuildInitialWerkerTask(buildConf, taskBuild.Hash, token, taskBuild.Branch, taskBuild.AcctRepo, taskBuild.By, nil, handler.GetVcsType())
		task.ChangesetData = &pb.ChangesetData{Branch: taskBuild.Branch, SubscriptionAlias: taskBuild.Subscription.Alias}
		if queueError := m.Signaler.OcyValidator.ValidateViability(task.Branch, task.BuildConf.Branches, nil, false); queueError != nil {
			logWithFields.WithField("error", queueError).Info("not queuing! this is fine, just doesn't fit requirements")
			return queueError
		}
		buildId, err := m.Signaler.CheckTaskAndStore(task)
		if err != nil {
			logWithFields.WithField("error", err).Error("unable to check task and store")
			return err
		}
		if task.SignaledBy == pb.SignaledBy_SUBSCRIBED {
			if err = m.Store.InsertSubscriptionData(taskBuild.Subscription.GetBuildId(), buildId, taskBuild.Subscription.GetActiveSubscriptionId()); err != nil {
				logWithFields.WithField("error", err).Error("unable to insert into active subscriptions table")
				// todo: not sure if we actually want to return here?
				return err
			}
		}
		if err = m.Signaler.QueueIfGoodConfig(buildId, task); err != nil {
			logWithFields.Error("unable to check task and store")
			return err
		}
		logWithFields.Info("successfully generated a werker task and added it to the werker queue for building.")
	default:
		err = errors.New("only supported topic is 'taskbuilder'")
		log.IncludeErrField(err).Error()
	}
	return err
}

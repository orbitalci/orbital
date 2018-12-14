package werkerinterface

import (
	"github.com/level11consulting/orbitalci/build/buildeventhandler/push/buildjob"
	"github.com/level11consulting/orbitalci/models"
	"github.com/level11consulting/orbitalci/models/pb"
)

//go:generate mockgen -source werkerinterface.go -destination werkerteller.mock.go -package werkerinterface

type CommitPushWerkerTeller interface {
	TellWerker(push *pb.Push, conf *buildjob.Signaler, handler models.VCSHandler, token string, force bool, sigBy pb.SignaledBy) error
}

type PRWerkerTeller interface {
	TellWerker(push *pb.PullRequest, prData *pb.PrWerkerData, conf *buildjob.Signaler, handler models.VCSHandler, token string, force bool, sigBy pb.SignaledBy) error
}

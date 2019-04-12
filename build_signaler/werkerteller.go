package build_signaler

import (
	"github.com/level11consulting/ocelot/build/eventhandler/push/buildjob"
	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/models/pb"
)

//go:generate mockgen -source werkerteller.go -destination werkerteller.mock.go -package build_signaler

type CommitPushWerkerTeller interface {
	TellWerker(push *pb.Push, conf *buildjob.Signaler, handler models.VCSHandler, token string, force bool, sigBy pb.SignaledBy) error
}

type PRWerkerTeller interface {
	TellWerker(push *pb.PullRequest, prData *pb.PrWerkerData, conf *buildjob.Signaler, handler models.VCSHandler, token string, force bool, sigBy pb.SignaledBy) error
}

package build_signaler

import (
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
)

//go:generate mockgen -source werkerteller.go -destination werkerteller.mock.go -package build_signaler

type CommitPushWerkerTeller interface {
	TellWerker(push *pb.Push, conf *Signaler, handler models.VCSHandler, token string, force bool, sigBy pb.SignaledBy) error
}

type PRWerkerTeller interface {
	TellWerker(push *pb.PullRequest, prData *pb.PrWerkerData, conf *Signaler, handler models.VCSHandler, token string, force bool, sigBy pb.SignaledBy) error
}

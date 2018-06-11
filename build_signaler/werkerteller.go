package build_signaler

import (
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
)


// made this interface for easy testing
type WerkerTeller interface {
	TellWerker(lastCommit string, conf *Signaler, branch string, handler models.VCSHandler, token, acctRepo string, commits []*pb.Commit, force bool, sigBy pb.SignaledBy, prData *pb.PullRequest) error
}

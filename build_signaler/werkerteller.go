package build_signaler

import (
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/models"
)


// made this interface for easy testing
type WerkerTeller interface {
	TellWerker(lastCommit string, conf *Signaler, branch string, handler models.VCSHandler, token, acctRepo string, viableData *build.Viable) (err error)
}

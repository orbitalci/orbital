package admin

import (

	"github.com/level11consulting/ocelot/build/vcshandler"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/level11consulting/ocelot/build/buildeventhandler/push/buildjob"
	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/shankj3/go-til/log"
)

var (
	triggeredBuilds = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_triggered_builds",
		Help: "builds triggered by a call to admin",
	}, []string{"account", "repository"})
)

func init() {
	prometheus.MustRegister(triggeredBuilds)
}

// getHandler returns a grpc status.Error
func (g *OcelotServerAPI) GetHandler(cfg *pb.VCSCreds) (models.VCSHandler, string, error) {
	if g.handler != nil {
		return g.handler, "token", nil
	}
	handler, token, err := vcshandler.GetHandler(cfg)
	if err != nil {
		log.IncludeErrField(err).Error()
		return nil, token, status.Errorf(codes.Internal, "Unable to retrieve the bitbucket client config for %s. \n Error: %s", cfg.AcctName, err.Error())
	}
	return handler, token, nil
}

func (g *OcelotServerAPI) GetSignaler() *buildjob.Signaler {
	return buildjob.NewSignaler(g.RemoteConfig, g.Deserializer, g.Producer, g.OcyValidator, g.Storage)
}
package admin

import (

	"github.com/level11consulting/ocelot/build/vcshandler"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/shankj3/go-til/log"
)

// getHandler returns a grpc status.Error
func (g *OcelotServerAPI) GetHandler(cfg *pb.VCSCreds) (models.VCSHandler, string, error) {
	if g.DeprecatedHandler.handler != nil {
		return g.DeprecatedHandler.handler, "token", nil
	}
	handler, token, err := vcshandler.GetHandler(cfg)
	if err != nil {
		log.IncludeErrField(err).Error()
		return nil, token, status.Errorf(codes.Internal, "Unable to retrieve the bitbucket client config for %s. \n Error: %s", cfg.AcctName, err.Error())
	}
	return handler, token, nil
}
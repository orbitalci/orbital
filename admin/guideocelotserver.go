package admin

import (
	"github.com/shankj3/ocelot/admin/models"
	"github.com/shankj3/ocelot/util"
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/shankj3/ocelot/util/ocelog"
	"github.com/shankj3/ocelot/util/deserialize"
)

//this is our grpc server struct
type guideOcelotServer struct {
	RemoteConfig	*util.RemoteConfig
	Deserializer	*deserialize.Deserializer
	AdminValidator	*AdminValidator
}

//TODO: what about adding error field to response? Do something nice about
func (g *guideOcelotServer) GetCreds(ctx context.Context, msg *empty.Empty) (*models.CredWrapper, error) {
	ocelog.Log().Debug("well at least we made it in teheheh")
	credWrapper := &models.CredWrapper{}

	creds, err := g.RemoteConfig.GetCredAt(util.ConfigPath, true)
	if err != nil {
		return credWrapper, err
	}

	for _, v := range creds {
		credWrapper.Credentials = append(credWrapper.Credentials, v)
	}
	return credWrapper, nil
}

func (g *guideOcelotServer) SetCreds(ctx context.Context, credentials *models.Credentials) (*empty.Empty, error) {
	err := g.AdminValidator.ValidateConfig(credentials)
	if err != nil {
		return nil, err
	}
	err = SetupCredentials(g, credentials)
	return nil, err
}

func NewGuideOcelotServer(config *util.RemoteConfig, d *deserialize.Deserializer, adminV *AdminValidator) models.GuideOcelotServer {
	guideOcelotServer := new(guideOcelotServer)
	guideOcelotServer.RemoteConfig = config
	guideOcelotServer.Deserializer = d
	guideOcelotServer.AdminValidator = adminV
	return guideOcelotServer
}
package admin

import (
	"bitbucket.org/level11consulting/go-til/deserialize"
	"bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"context"
	"github.com/golang/protobuf/ptypes/empty"
)

//this is our grpc server struct
type guideOcelotServer struct {
	RemoteConfig   *cred.RemoteConfig
	Deserializer   *deserialize.Deserializer
	AdminValidator *AdminValidator
}

//TODO: what about adding error field to response? Do something nice about
func (g *guideOcelotServer) GetCreds(ctx context.Context, msg *empty.Empty) (*models.CredWrapper, error) {
	log.Log().Debug("well at least we made it in teheheh")
	credWrapper := &models.CredWrapper{}

	creds, err := g.RemoteConfig.GetCredAt(cred.ConfigPath, true)
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
	return &empty.Empty{}, err
}

func NewGuideOcelotServer(config *cred.RemoteConfig, d *deserialize.Deserializer, adminV *AdminValidator) models.GuideOcelotServer {
	guideOcelotServer := new(guideOcelotServer)
	guideOcelotServer.RemoteConfig = config
	guideOcelotServer.Deserializer = d
	guideOcelotServer.AdminValidator = adminV
	return guideOcelotServer
}
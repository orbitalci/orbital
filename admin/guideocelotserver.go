package main

import (
	"github.com/shankj3/ocelot/admin/models"
	"github.com/shankj3/ocelot/util"
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/shankj3/ocelot/util/ocelog"
)

//this is our grpc server struct
type guideOcelotServer struct {
	remoteConfig	*util.RemoteConfig
}

func (g *guideOcelotServer) GetCreds(ctx context.Context, msg *empty.Empty) (*models.CredWrapper, error) {
	ocelog.Log().Debug("well at least we made it in teheheh")
	credWrapper := &models.CredWrapper{}

	//creds, err := g.remoteConfig.GetCredAt(util.ConfigPath, true)
	//if err != nil {
	//	return credWrapper, err
	//}
	//
	//for _, v := range creds {
	//	credWrapper.Credentials = append(credWrapper.Credentials, v)
	//}
	return credWrapper, nil
}

func NewGuideOcelotServer() models.GuideOcelotServer {
	return new(guideOcelotServer)
}
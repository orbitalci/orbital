package main

import (
	"github.com/shankj3/ocelot/admin/models"
	"github.com/shankj3/ocelot/util"
	"context"
	"github.com/golang/protobuf/ptypes/empty"
)

//this is our grpc server struct
type guideOcelotServer struct {
	remoteConfig	*util.RemoteConfig
	//TODO: probably want other properties here
}

func (g *guideOcelotServer) GetCreds(ctx context.Context, msg *empty.Empty) (*models.CredWrapper, error) {
	credWrapper := &models.CredWrapper{}

	creds, err := g.remoteConfig.GetCredAt(util.ConfigPath, true)
	if err != nil {
		return credWrapper, err
	}

	for _, v := range creds {
		credWrapper.Credentials = append(credWrapper.Credentials, v)
	}
	return credWrapper, nil
}

func NewGuideOcelotServer() models.GuideOcelotServer {
	return new(guideOcelotServer)
}
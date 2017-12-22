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
	RepoValidator  *RepoValidator
}

//TODO: what about adding error field to response? Do something nice about
func (g *guideOcelotServer) GetVCSCreds(ctx context.Context, msg *empty.Empty) (*models.CredWrapper, error) {
	log.Log().Debug("well at least we made it in teheheh")
	credWrapper := &models.CredWrapper{}
	creds, err := g.RemoteConfig.GetCredAt(cred.VCSPath, true, cred.Vcs)
	if err != nil {
		return credWrapper, err
	}

	for _, v := range creds {
		credWrapper.Vcs = append(credWrapper.Vcs, v.(*models.VCSCreds))
	}
	return credWrapper, nil
}

// for checking if the server is reachable
func (g *guideOcelotServer) CheckConn(ctx context.Context, msg *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}


func (g *guideOcelotServer) SetVCSCreds(ctx context.Context, credentials *models.VCSCreds) (*empty.Empty, error) {
	err := g.AdminValidator.ValidateConfig(credentials)
	if err != nil {
		return nil, err
	}
	err = SetupCredentials(g, credentials)
	return &empty.Empty{}, err
}


func (g *guideOcelotServer) GetRepoCreds(ctx context.Context, msg *empty.Empty) (*models.RepoCredWrapper, error) {
	credWrapper := &models.RepoCredWrapper{}
	creds, err := g.RemoteConfig.GetCredAt(cred.RepoPath, true, cred.Repo)
	if err != nil {
		return credWrapper, err
	}
	for _, v := range creds {
		credWrapper.Repo = append(credWrapper.Repo, v.(*models.RepoCreds))
	}
	return credWrapper, nil
}

func (g *guideOcelotServer) SetRepoCreds(ctx context.Context, creds *models.RepoCreds) (*empty.Empty, error) {
	err := g.RepoValidator.ValidateConfig(creds)
	if err != nil {
		return nil, err
	}
	err = SetupRepoCredentials(g, creds)
	return &empty.Empty{}, err
}

func (g *guideOcelotServer) GetAllCreds(ctx context.Context, msg *empty.Empty) (*models.AllCredsWrapper, error) {
	allCreds := &models.AllCredsWrapper{}
	repoCreds, err := g.GetRepoCreds(ctx, msg)
	if err != nil {
		return allCreds, err
	}
	allCreds.RepoCreds = repoCreds
	adminCreds, err := g.GetVCSCreds(ctx, msg)
	if err != nil {
		return allCreds, err
	}
	allCreds.VcsCreds = adminCreds
	return allCreds, nil
}


func NewGuideOcelotServer(config *cred.RemoteConfig, d *deserialize.Deserializer, adminV *AdminValidator, repoV *RepoValidator) models.GuideOcelotServer {
	guideOcelotServer := new(guideOcelotServer)
	guideOcelotServer.RemoteConfig = config
	guideOcelotServer.Deserializer = d
	guideOcelotServer.AdminValidator = adminV
	guideOcelotServer.RepoValidator = repoV
	return guideOcelotServer
}
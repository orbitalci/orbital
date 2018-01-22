package admin

import (
	"bitbucket.org/level11consulting/go-til/deserialize"
	"bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	rt "bitbucket.org/level11consulting/ocelot/util/buildruntime"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//this is our grpc server struct
type guideOcelotServer struct {
	RemoteConfig   cred.CVRemoteConfig
	Deserializer   *deserialize.Deserializer
	AdminValidator *AdminValidator
	RepoValidator  *RepoValidator
	Storage 	   storage.BuildOutputStorage
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
		return nil, status.Errorf(codes.FailedPrecondition, "failed vcs creds validation! error: %s", err.Error())
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
		return nil, status.Errorf(codes.FailedPrecondition, "failed repo creds validation! error: %s", err.Error())
	}
	err = SetupRepoCredentials(g, creds)
	return &empty.Empty{}, err
}

func (g *guideOcelotServer) GetAllCreds(ctx context.Context, msg *empty.Empty) (*models.AllCredsWrapper, error) {
	allCreds := &models.AllCredsWrapper{}
	repoCreds, err := g.GetRepoCreds(ctx, msg)
	if err != nil {
		return allCreds, status.Errorf(codes.Internal, "unable to get repo creds! error: %s", err.Error())
	}
	allCreds.RepoCreds = repoCreds
	adminCreds, err := g.GetVCSCreds(ctx, msg)
	if err != nil {
		return allCreds, status.Errorf(codes.Internal, "unable to get vcs creds! error: %s", err.Error())
	}
	allCreds.VcsCreds = adminCreds
	return allCreds, nil
}

func (g *guideOcelotServer) BuildRuntime(ctx context.Context, bq *models.BuildQuery) (*models.BuildRuntimeInfo, error) {
	buildRtInfo, err := rt.GetBuildRuntime(g.RemoteConfig.GetConsul(), bq.Hash)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &models.BuildRuntimeInfo{Done: buildRtInfo.Done, Ip: buildRtInfo.Ip, GrpcPort: buildRtInfo.GrpcPort}, nil
}

// todo: calling Logs should stream from build storage, this is doing nothing
func (g *guideOcelotServer) Logs(bq *models.BuildQuery, stream models.GuideOcelot_LogsServer) error {
	if !rt.CheckIfBuildDone(g.RemoteConfig.GetConsul(), bq.Hash) {
		stream.Send(&models.LogResponse{OutputLine: "build is not finished, use BuildRuntime method and stream from the werker registered",})
	} else {
		data, err := g.Storage.Retrieve(bq.Hash)
		if err != nil {
			return status.Error(codes.DataLoss, fmt.Sprintf("Unable to retrieve from %s. \nError: %s", g.Storage.StorageType(), err.Error()))
		}
		scanner := bufio.NewScanner(bytes.NewReader(data))
		for scanner.Scan() {
			resp := &models.LogResponse{OutputLine: scanner.Text()}
			stream.Send(resp)
		}
	}
	return nil
}


func NewGuideOcelotServer(config cred.CVRemoteConfig, d *deserialize.Deserializer, adminV *AdminValidator, repoV *RepoValidator, storage storage.BuildOutputStorage) models.GuideOcelotServer {
	// changing to this style of instantiation cuz thread safe (idk read it on some best practices, it just looks
	// purdier to me anyway
	guideOcelotServer := &guideOcelotServer{
		RemoteConfig: config,
		Deserializer: d,
		AdminValidator: adminV,
		RepoValidator: repoV,
		Storage: storage,
	}
	return guideOcelotServer
}
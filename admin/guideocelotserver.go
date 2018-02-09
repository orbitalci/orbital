package admin

import (
	"bitbucket.org/level11consulting/go-til/deserialize"
	"bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	rt "bitbucket.org/level11consulting/ocelot/util/buildruntime"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	md "bitbucket.org/level11consulting/ocelot/util/storage/models"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//this is our grpc server struct
type guideOcelotServer struct {
	RemoteConfig   cred.CVRemoteConfig
	Deserializer   *deserialize.Deserializer
	AdminValidator *AdminValidator
	RepoValidator  *RepoValidator
	Storage 	   storage.OcelotStorage
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

func (g *guideOcelotServer) BuildRuntime(ctx context.Context, bq *models.BuildQuery) (*models.Builds, error) {
	//find matching hashes in consul
	buildRtInfo, err := rt.GetBuildRuntime(g.RemoteConfig.GetConsul(), bq.Hash)
	if err != nil {
		if _, ok := err.(*rt.ErrBuildDone); !ok {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	//add matching hashes in db if exists and add acctname/repo to ones found in consul
	dbResults, err := g.Storage.RetrieveHashStartsWith(bq.Hash)
	for _, build := range dbResults {
		if _, ok := buildRtInfo[build.Hash]; !ok {
			buildRtInfo[build.Hash] = &models.BuildRuntimeInfo{
				Hash: build.Hash,
				Done: true,
			}
		}
		buildRtInfo[build.Hash].AcctName = build.Account
		buildRtInfo[build.Hash].RepoName = build.Repo
	}

	builds := &models.Builds{
		Builds : buildRtInfo,
	}
	return builds, err
}


func (g *guideOcelotServer) Logs(bq *models.BuildQuery, stream models.GuideOcelot_LogsServer) error {
	if !rt.CheckIfBuildDone(g.RemoteConfig.GetConsul(), g.Storage, bq.Hash) {
		stream.Send(&models.LogResponse{OutputLine: "build is not finished, use BuildRuntime method and stream from the werker registered",})
	} else {
		var out md.BuildOutput
		var err error
		if bq.BuildId != 0 {
			out, err = g.Storage.RetrieveOut(bq.BuildId)
		} else {
			out, err = g.Storage.RetrieveLastOutByHash(bq.Hash)
		}
		if err != nil {
			return status.Error(codes.DataLoss, fmt.Sprintf("Unable to retrieve from %s. \nError: %s", g.Storage.StorageType(), err.Error()))
		}
		scanner := bufio.NewScanner(bytes.NewReader(out.Output))
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)
		for scanner.Scan() {
			resp := &models.LogResponse{OutputLine: scanner.Text()}
			stream.Send(resp)
		}
		if err := scanner.Err(); err != nil {
			log.IncludeErrField(err).Error("error encountered scanning from " + g.Storage.StorageType())
			return status.Error(codes.DataLoss, fmt.Sprintf("Error was encountered while sending data from %s. \nError: %s", g.Storage.StorageType(), err.Error()))
		}
	}
	return nil
}

func (g *guideOcelotServer) LastFewSummaries(ctx context.Context, repoAct *models.RepoAccount) (*models.Summaries, error) {
	var summaries = &models.Summaries{}
	modelz, err := g.Storage.RetrieveLastFewSums(repoAct.Repo, repoAct.Account, repoAct.Limit)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	for _, model := range modelz {
		summary := &models.BuildSummary{
			Hash: model.Hash,
			Failed: model.Failed,
			BuildTime: &timestamp.Timestamp{Seconds: model.BuildTime.UTC().Unix()},
			Account: model.Account,
			BuildDuration: model.BuildDuration,
			Repo: model.Repo,
			Branch: model.Branch,
			BuildId: model.BuildId,
		}
		summaries.Sums = append(summaries.Sums, summary)
	}
	return summaries, nil

}

func NewGuideOcelotServer(config cred.CVRemoteConfig, d *deserialize.Deserializer, adminV *AdminValidator, repoV *RepoValidator, storage storage.OcelotStorage) models.GuideOcelotServer {
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
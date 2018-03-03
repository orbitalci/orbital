package admin

import (
	"bitbucket.org/level11consulting/go-til/deserialize"
	"bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/go-til/net"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	rt "bitbucket.org/level11consulting/ocelot/util/buildruntime"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/handler"
	"bitbucket.org/level11consulting/ocelot/util/storage"
	md "bitbucket.org/level11consulting/ocelot/util/storage/models"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

//this is our grpc server, it responds to client requests
type guideOcelotServer struct {
	RemoteConfig   cred.CVRemoteConfig
	Deserializer   *deserialize.Deserializer
	AdminValidator *AdminValidator
	RepoValidator  *RepoValidator
	Storage        storage.OcelotStorage
	Producer       *nsqpb.PbProduce
}

func (g *guideOcelotServer) BuildRepoAndHash(ctx context.Context, buildReq *models.AcctRepoAndHash) (*models.BuildSummary, error) {
	if buildReq == nil || len(buildReq.AcctRepo) == 0 || len(buildReq.PartialHash) == 0{
		return nil, errors.New("please pass a valid account/repo_name and hash")
	}
	go g.Producer.WriteProto(buildReq, "build_please")
	return &models.BuildSummary{}, nil
}


func (g *guideOcelotServer) GetVCSCreds(ctx context.Context, msg *empty.Empty) (*models.CredWrapper, error) {
	credWrapper := &models.CredWrapper{}
	vcs := models.NewVCSCreds()
	creds, err := g.RemoteConfig.GetCredAt(cred.VCSPath, true, vcs)
	if err != nil {
		return credWrapper, err
	}

	for _, v := range creds {
		vcsCred := v.(*models.VCSCreds)
		sshKeyPath := cred.BuildCredPath(vcsCred.Type, vcsCred.AcctName, cred.Vcs)
		err := g.RemoteConfig.CheckSSHKeyExists(sshKeyPath)
		if err != nil {
			vcsCred.SshFileLoc = "\033[0;33mNo SSH Key\033[0m"
		} else {
			vcsCred.SshFileLoc = "\033[0;34mSSH Key on file\033[0m"
		}
		credWrapper.Vcs = append(credWrapper.Vcs, vcsCred)
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
	repo := models.NewRepoCreds()
	creds, err := g.RemoteConfig.GetCredAt(cred.RepoPath, true, repo)
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
	buildRtInfo := make(map[string]*models.BuildRuntimeInfo)
	var err error

	if len(bq.Hash) > 0 {
		//find matching hashes in consul by git hash
		buildRtInfo, err = rt.GetBuildRuntime(g.RemoteConfig.GetConsul(), bq.Hash)
		if err != nil {
			if _, ok := err.(*rt.ErrBuildDone); !ok {
				log.IncludeErrField(err)
				return nil, status.Error(codes.Internal, err.Error())
			} else {
				//we set error back to nil so that we can continue with the rest of the logic here
				err = nil
			}
		}

		//add matching hashes in db if exists and add acctname/repo to ones found in consul
		dbResults, err := g.Storage.RetrieveHashStartsWith(bq.Hash)

		if err != nil {
			return &models.Builds{
				Builds: buildRtInfo,
			}, err
		}

		for _, build := range dbResults {
			if _, ok := buildRtInfo[build.Hash]; !ok {
				buildRtInfo[build.Hash] = &models.BuildRuntimeInfo{
					Hash: build.Hash,
					// if a result was found in the database but not in GetBuildRuntime, the build is done
					Done: true,
				}
			}
			buildRtInfo[build.Hash].AcctName = build.Account
			buildRtInfo[build.Hash].RepoName = build.Repo
		}
	}

	//if a valid build id passed, go ask db for entries
	if bq.BuildId > 0 {
		buildSum, err := g.Storage.RetrieveSumByBuildId(bq.BuildId)
		if err != nil {
			return &models.Builds{
				Builds: buildRtInfo,
			}, err
		}

		buildRtInfo[buildSum.Hash] = &models.BuildRuntimeInfo{
			Hash:     buildSum.Hash,
			Done:     true,
			AcctName: buildSum.Account,
			RepoName: buildSum.Repo,
		}
	}

	builds := &models.Builds{
		Builds: buildRtInfo,
	}
	return builds, err
}

func (g *guideOcelotServer) Logs(bq *models.BuildQuery, stream models.GuideOcelot_LogsServer) error {
	if !rt.CheckIfBuildDone(g.RemoteConfig.GetConsul(), g.Storage, bq.Hash) {
		stream.Send(&models.LogResponse{OutputLine: "build is not finished, use BuildRuntime method and stream from the werker registered"})
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
			Hash:          model.Hash,
			Failed:        model.Failed,
			BuildTime:     &timestamp.Timestamp{Seconds: model.BuildTime.UTC().Unix()},
			Account:       model.Account,
			BuildDuration: model.BuildDuration,
			Repo:          model.Repo,
			Branch:        model.Branch,
			BuildId:       model.BuildId,
		}
		summaries.Sums = append(summaries.Sums, summary)
	}
	return summaries, nil

}

func (g *guideOcelotServer) WatchRepo(ctx context.Context, repoAcct *models.RepoAccount) (*empty.Empty, error) {
	var vcs *models.VCSCreds

	bbCreds, err := g.RemoteConfig.GetCredAt(cred.BuildCredPath("bitbucket", repoAcct.Account, cred.Vcs), false, vcs)
	if err != nil {
		return &empty.Empty{}, err
	}

	//TODO: what do we even do if there's more than one?
	for _, v := range bbCreds {
		vcs = v.(*models.VCSCreds)
		bbClient := &net.OAuthClient{}
		bbClient.Setup(vcs)

		bbHandler := handler.GetBitbucketHandler(vcs, bbClient)

		repoDetail, err := bbHandler.GetRepoDetail(fmt.Sprintf("%s/%s", repoAcct.Account, repoAcct.Repo))
		if err != nil {
			return nil, err
		}

		webhookURL := repoDetail.GetLinks().GetHooks().GetHref()
		err = bbHandler.CreateWebhook(webhookURL)

		if err != nil {
			return nil, err
		}
		return &empty.Empty{}, nil
	}
	return &empty.Empty{}, nil
}

//StatusByHash will retrieve you the status (build summary + stages) of a partial git hash
func (g *guideOcelotServer) GetStatus(ctx context.Context, query *models.StatusQuery) (*models.Status, error) {
	//hash first
	if len(query.Hash) > 0 {
		partialHash := query.Hash
		buildSum, err := g.Storage.RetrieveLatestSum(partialHash)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		stageResults, err := g.Storage.RetrieveStageDetail(buildSum.BuildId)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		result := ParseStagesByBuildId(buildSum, stageResults)
		return result, nil
	}

	if len(query.AcctName) > 0 && len(query.RepoName) > 0 {
		buildSums, err := g.Storage.RetrieveLastFewSums(query.RepoName, query.AcctName, 1)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		if len(buildSums) == 1 {
			buildSum := buildSums[0]

			stageResults, err := g.Storage.RetrieveStageDetail(buildSum.BuildId)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			result := ParseStagesByBuildId(buildSum, stageResults)
			return result, nil
		} else {
			uhOh := errors.New(fmt.Sprintf("there is no ONE entry that matches the acctname/repo %s/%s", query.AcctName, query.RepoName))
			log.IncludeErrField(uhOh)
			return nil, status.Error(codes.Internal, uhOh.Error())
		}
	}

	if len(query.PartialRepo) > 0 {
		buildSums, err := g.Storage.RetrieveAcctRepo(strings.TrimSpace(query.PartialRepo))
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		if len(buildSums) == 1 {
			buildSum, err := g.Storage.RetrieveLastFewSums(buildSums[0].Repo, buildSums[0].Account, 1)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}

			stageResults, err := g.Storage.RetrieveStageDetail(buildSum[0].BuildId)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			result := ParseStagesByBuildId(buildSum[0], stageResults)
			return result, nil
		} else {
			var uhOh error
			if len(buildSums) == 0 {
				uhOh = errors.New(fmt.Sprintf("there are no repositories starting with %s", query.PartialRepo))
			} else {
				var matches []string
				for _, buildSum := range buildSums {
					matches = append(matches, buildSum.Account+"/"+buildSum.Repo)
				}
				uhOh = errors.New(fmt.Sprintf("there are %v repositories starting with %s: %s", len(buildSums), query.PartialRepo, strings.Join(matches, ",")))
			}
			log.IncludeErrField(uhOh)
			return nil, status.Error(codes.Internal, uhOh.Error())
		}
	}

	return &models.Status{}, nil
}

func (g *guideOcelotServer) SetVCSPrivateKey(ctx context.Context, sshKeyWrapper *models.SSHKeyWrapper) (*empty.Empty, error) {
	sshKeyPath := cred.BuildCredPath(sshKeyWrapper.Type, sshKeyWrapper.AcctName, cred.Vcs)
	err := g.RemoteConfig.AddSSHKey(sshKeyPath, sshKeyWrapper.PrivateKey)
	if err != nil {
		return &empty.Empty{}, err
	}
	return &empty.Empty{}, nil
}

func NewGuideOcelotServer(config cred.CVRemoteConfig, d *deserialize.Deserializer, adminV *AdminValidator, repoV *RepoValidator, storage storage.OcelotStorage) models.GuideOcelotServer {
	// changing to this style of instantiation cuz thread safe (idk read it on some best practices, it just looks
	// purdier to me anyway
	guideOcelotServer := &guideOcelotServer{
		RemoteConfig:   config,
		Deserializer:   d,
		AdminValidator: adminV,
		RepoValidator:  repoV,
		Storage:        storage,
		Producer:       nsqpb.GetInitProducer(),
	}
	return guideOcelotServer
}

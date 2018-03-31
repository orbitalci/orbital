package admin

import (
	"bitbucket.org/level11consulting/go-til/deserialize"
	"bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/go-til/net"
	"bitbucket.org/level11consulting/go-til/nsqpb"
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/util/build"
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
	OcyValidator   *build.OcelotValidator
	Storage        storage.OcelotStorage
	Producer       *nsqpb.PbProduce
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

func (g *guideOcelotServer) FindWerker(ctx context.Context, br *models.BuildReq) (*models.BuildRuntimeInfo, error) {
	if len(br.Hash) > 0 {
		//find matching hashes in consul by git hash
		buildRtInfo, err := rt.GetBuildRuntime(g.RemoteConfig.GetConsul(), br.Hash)
		if err != nil {
			if _, ok := err.(*rt.ErrBuildDone); !ok {
				return nil, err
			}
		}

		if len(buildRtInfo) == 0 || len(buildRtInfo) > 1 {
			return nil, errors.New("ONE and ONE ONLY match should be found for your hash")
		}

		for _, v := range buildRtInfo {
			return v, nil
		}
	} else {
		return nil, errors.New("Please pass a hash")
	}
	return nil, nil
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
			}, handleStorageError(err)
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
			}, handleStorageError(err)
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
		stream.Send(&models.LineResponse{OutputLine: "build is not finished, use BuildRuntime method and stream from the werker registered"})
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
			resp := &models.LineResponse{OutputLine: scanner.Text()}
			stream.Send(resp)
		}
		if err := scanner.Err(); err != nil {
			log.IncludeErrField(err).Error("error encountered scanning from " + g.Storage.StorageType())
			return status.Error(codes.DataLoss, fmt.Sprintf("Error was encountered while sending data from %s. \nError: %s", g.Storage.StorageType(), err.Error()))
		}
	}
	return nil
}

func (g *guideOcelotServer) BuildRepoAndHash(buildReq *models.BuildReq, stream models.GuideOcelot_BuildRepoAndHashServer) error {
	if buildReq == nil || len(buildReq.AcctRepo) == 0 || len(buildReq.Hash) == 0 {
		return errors.New("please pass a valid account/repo_name and hash")
	}

	stream.Send(RespWrap(fmt.Sprintf("Searching for VCS creds belonging to %s...", buildReq.AcctRepo)))
	cfg, err := build.GetVcsCreds(buildReq.AcctRepo, g.RemoteConfig)
	if err != nil {
		log.IncludeErrField(err).Error()
		return err
	}
	stream.Send(RespWrap(fmt.Sprintf("Successfully found VCS credentials belonging to %s %s", buildReq.AcctRepo, md.CHECKMARK)))
	stream.Send(RespWrap("Validating VCS Credentials..."))
	bbHandler, token, err := handler.GetBitbucketClient(cfg)
	if err != nil {
		log.IncludeErrField(err).Error()
		return status.Error(codes.Internal, fmt.Sprintf("Unable to retrieve the bitbucket client config for %s. \n Error: %s", buildReq.AcctRepo, err.Error()))
	}
	stream.Send(RespWrap(fmt.Sprintf("Successfully used VCS Credentials to obtain a token %s", md.CHECKMARK)))

	var branch string
	var fullHash string
	stream.Send(RespWrap(fmt.Sprintf("Looking in previous builds for %s...", buildReq.Hash)))
	buildSum, err := g.Storage.RetrieveLatestSum(buildReq.Hash)
	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); !ok {
			log.IncludeErrField(err).Error("could not retrieve latest build summary")
			return status.Error(codes.ResourceExhausted, fmt.Sprintf("Unable to connect to the database, therefore this operation is not available at this time."))
		}
		//at this point error must be because we couldn't find hash starting with query
		warnMsg := fmt.Sprintf("There are no previous builds starting with hash %s...", buildReq.Hash)
		log.IncludeErrField(err).Warning(warnMsg)
		stream.Send(RespWrap(warnMsg))

		if len(buildReq.Branch) == 0 {
			noBranchErr := errors.New("Branch is a required field if a previous build starting with the specified hash cannot be found. Please pass the branch flag and try again!")
			log.IncludeErrField(noBranchErr).Error("branch len is 0")
			return noBranchErr
		}

		fullHash = buildReq.Hash
		branch = buildReq.Branch
	} else {
		acct, repo := build.GetAcctRepo(buildReq.AcctRepo)
		if buildSum.Repo != repo || buildSum.Account != acct {
			mismatchErr := errors.New(fmt.Sprintf("The account/repo passed (%s) doesn't match with the account/repo (%s) associated with build #%v", buildReq.AcctRepo, buildSum.Account + "/" + buildSum.Repo, buildSum.BuildId))
			log.IncludeErrField(mismatchErr).Error()
			return mismatchErr
		}


		if len(buildReq.Branch) == 0 {
			stream.Send(RespWrap(fmt.Sprintf("No branch was passed, using `%s` from build #%v instead...", buildSum.Branch, buildSum.BuildId)))
			branch = buildSum.Branch
		} else {
			branch = buildReq.Branch
		}

		fullHash = buildSum.Hash
		stream.Send(RespWrap(fmt.Sprintf("Found a previous build starting with hash %s, now building branch %s %s", buildReq.Hash, branch, md.CHECKMARK)))
	}

	stream.Send(RespWrap(fmt.Sprintf("Retrieving ocelot.yml for %s...", buildReq.AcctRepo)))
	buildConf, _, err := build.GetBBConfig(g.RemoteConfig, buildReq.AcctRepo, fullHash, g.Deserializer, bbHandler)
	if err != nil {
		log.IncludeErrField(err).Error("couldn't get bb config")
		return err
	}
	stream.Send(RespWrap(fmt.Sprintf("Successfully retrieved ocelot.yml for %s %s", buildReq.AcctRepo, md.CHECKMARK)))
	stream.Send(RespWrap(fmt.Sprintf("Storing build data for %s...", buildReq.AcctRepo)))
	if err = build.QueueAndStore(fullHash, branch, buildReq.AcctRepo, token, g.RemoteConfig, buildConf, g.OcyValidator, g.Producer, g.Storage); err != nil {
		log.IncludeErrField(err).Error("couldn't add to build queue or store in db")
		return err
	}
	stream.Send(RespWrap(fmt.Sprintf("Build started for %s belonging to %s %s", fullHash, buildReq.AcctRepo, md.CHECKMARK)))
	return nil
}


func (g *guideOcelotServer) LastFewSummaries(ctx context.Context, repoAct *models.RepoAccount) (*models.Summaries, error) {
	var summaries = &models.Summaries{}
	modelz, err := g.Storage.RetrieveLastFewSums(repoAct.Repo, repoAct.Account, repoAct.Limit)
	if err != nil {
		return nil, handleStorageError(err)
	}
	if len(modelz) == 0 {
		return nil, status.Error(codes.NotFound, "no entries found")
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

	if bbCreds == nil || len(bbCreds) == 0 {
		return &empty.Empty{}, errors.New(fmt.Sprintf("could not find credentials belonging to %s", repoAcct.Account))
	}

	//TODO: what do we even do if there's more than one?
	for _, v := range bbCreds {
		vcs = v.(*models.VCSCreds)
		bbClient := &net.OAuthClient{}
		bbClient.Setup(vcs)

		bbHandler := handler.GetBitbucketHandler(vcs, bbClient)
		repoDetail, err := bbHandler.GetRepoDetail(fmt.Sprintf("%s/%s", repoAcct.Account, repoAcct.Repo))
		if repoDetail.Type == "error" || err != nil {
			return &empty.Empty{}, errors.New(fmt.Sprintf("could not get repository detail at %s/%s", repoAcct.Account, repoAcct.Repo))
		}

		webhookURL := repoDetail.GetLinks().GetHooks().GetHref()
		err = bbHandler.CreateWebhook(webhookURL)

		if err != nil {
			return &empty.Empty{}, err
		}
		return &empty.Empty{}, nil
	}

	return &empty.Empty{}, nil
}

//StatusByHash will retrieve you the status (build summary + stages) of a partial git hash
func (g *guideOcelotServer) GetStatus(ctx context.Context, query *models.StatusQuery) (result *models.Status, err error) {
	var buildSum md.BuildSummary
	if len(query.Hash) > 0 {
		partialHash := query.Hash
		buildSum, err = g.Storage.RetrieveLatestSum(partialHash)
		if err != nil {
			return nil, handleStorageError(err)
		}
		goto BUILD_FOUND

	}
	if len(query.AcctName) > 0 && len(query.RepoName) > 0 {
		buildSums, err := g.Storage.RetrieveLastFewSums(query.RepoName, query.AcctName, 1)
		if err != nil {
			return nil, handleStorageError(err)
		}
		if len(buildSums) == 1 {
			buildSum = buildSums[0]
			goto BUILD_FOUND
		} else if len(buildSums) == 0 {
			uhOh := errors.New(fmt.Sprintf("There are no entries that match the acctname/repo %s/%s", query.AcctName, query.RepoName))
			log.IncludeErrField(uhOh).Error()
			return nil, status.Error(codes.NotFound, uhOh.Error())
		} else {
			// todo: this is logging even when there isn't a match in the db, probably an issue with RetrieveLastFewSums not returning error if there are no rows
			uhOh := errors.New(fmt.Sprintf("there is no ONE entry that matches the acctname/repo %s/%s", query.AcctName, query.RepoName))
			log.IncludeErrField(uhOh)
			return nil, status.Error(codes.Internal, uhOh.Error())
		}
	}

	if len(query.PartialRepo) > 0 {
		buildSums, err := g.Storage.RetrieveAcctRepo(strings.TrimSpace(query.PartialRepo))
		if err != nil {
			return nil, handleStorageError(err)
		}

		if len(buildSums) == 1 {
			buildSumz, err := g.Storage.RetrieveLastFewSums(buildSums[0].Repo, buildSums[0].Account, 1)
			if err != nil {
				return nil, handleStorageError(err)
			}
			buildSum = buildSumz[0]
			goto BUILD_FOUND
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
	return
BUILD_FOUND:
	stageResults, err := g.Storage.RetrieveStageDetail(buildSum.BuildId)
	if err != nil {
		return nil, handleStorageError(err)
	}
	result = ParseStagesByBuildId(buildSum, stageResults)
	inConsul, err := rt.CheckBuildInConsul(g.RemoteConfig.GetConsul(), buildSum.Hash)
	if err != nil {
		return nil, status.Error(codes.Internal, "An error occurred checking build status in consul. Cannot retrieve status at this time.\n\n" + err.Error())
	}
	result.IsInConsul = inConsul
	return
}

func (g *guideOcelotServer) SetVCSPrivateKey(ctx context.Context, sshKeyWrapper *models.SSHKeyWrapper) (*empty.Empty, error) {
	sshKeyPath := cred.BuildCredPath(sshKeyWrapper.Type, sshKeyWrapper.AcctName, cred.Vcs)
	err := g.RemoteConfig.AddSSHKey(sshKeyPath, sshKeyWrapper.PrivateKey)
	if err != nil {
		return &empty.Empty{}, err
	}
	return &empty.Empty{}, nil
}

func (g *guideOcelotServer) PollRepo(ctx context.Context, poll *models.PollRequest) (*empty.Empty, error) {
	log.Log().Info("recieved poll request for ", poll.Account, poll.Repo, poll.Cron)
	empti := &empty.Empty{}
	exists, err := g.Storage.PollExists(poll.Account, poll.Repo)
	if err != nil {
		return empti, status.Error(codes.Internal, err.Error())
	}
	if exists == true {
		log.Log().Info("updating poll in db")
		if err = g.Storage.UpdatePoll(poll.Account, poll.Repo, poll.Cron, poll.Branches); err != nil {
			msg := "unable to update poll in storage"
			log.IncludeErrField(err).Error(msg)
			return empti, status.Error(codes.Internal, msg + ": " + err.Error())
		}
	} else {
		log.Log().Info("inserting poll in db")
		if err = g.Storage.InsertPoll(poll.Account, poll.Repo, poll.Cron, poll.Branches); err != nil {
			msg := "unable to insert poll into storage"
			log.IncludeErrField(err).Error(msg)
			return empti, status.Error(codes.Internal, msg + ": " + err.Error())
		}
	}
	log.Log().WithField("account", poll.Account).WithField("repo", poll.Repo).Info("successfully added/updated poll in storage")
	err = g.Producer.WriteProto(poll, "poll_please")
	if err != nil {
		log.IncludeErrField(err).Error("couldn't write to queue producer at poll_please")
		return empti, status.Error(codes.Internal, err.Error())
	}
	return empti, nil
}

func (g *guideOcelotServer) DeletePollRepo(ctx context.Context, poll *models.PollRequest) (*empty.Empty, error) {
	log.Log().Info("received delete poll request for ", poll.Account, " ", poll.Repo)
	empti := &empty.Empty{}
	if err := g.Storage.DeletePoll(poll.Account, poll.Repo); err != nil {
		log.IncludeErrField(err).WithField("account", poll.Account).WithField("repo", poll.Repo).Error("couldn't delete poll")
	}
	log.Log().WithField("account", poll.Account).WithField("repo", poll.Repo).Info("successfully deleted poll in storage")
	if err := g.Producer.WriteProto(poll, "no_poll_please"); err != nil {
		log.IncludeErrField(err).Error("couldn't write to queue producer at no_poll_please")
		return empti, status.Error(codes.Internal, err.Error())
	}
	return empti, nil
}

// todo: add acct/repo action later
func (g *guideOcelotServer) ListPolledRepos(context.Context, *empty.Empty) (*models.Polls, error) {
	polls, err := g.Storage.GetAllPolls()
	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); !ok {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	pollz := &models.Polls{}
	for _, pll := range polls {
		pbpoll := &models.PollRequest{
			Account: pll.Account,
			Repo: pll.Repo,
			Cron: pll.Cron,
			Branches: pll.Branches,
			LastCronTime: &timestamp.Timestamp{Seconds:pll.LastCron.Unix(), Nanos:0},
		}
		pollz.Polls = append(pollz.Polls, pbpoll)
	}
	return pollz, nil
}

func NewGuideOcelotServer(config cred.CVRemoteConfig, d *deserialize.Deserializer, adminV *AdminValidator, repoV *RepoValidator, storage storage.OcelotStorage) models.GuideOcelotServer {
	// changing to this style of instantiation cuz thread safe (idk read it on some best practices, it just looks
	// purdier to me anyway
	guideOcelotServer := &guideOcelotServer{
		OcyValidator:   build.GetOcelotValidator(),
		RemoteConfig:   config,
		Deserializer:   d,
		AdminValidator: adminV,
		RepoValidator:  repoV,
		Storage:        storage,
		Producer:       nsqpb.GetInitProducer(),
	}
	return guideOcelotServer
}

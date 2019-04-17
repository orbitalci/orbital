package admin

import (
	"context"
	"errors"
	"fmt"

	"github.com/level11consulting/ocelot/build/buildeventhandler/push/buildjob"
	"github.com/level11consulting/ocelot/build/helpers/buildscript/download"
	stringbuilder "github.com/level11consulting/ocelot/build/helpers/stringbuilder/accountrepo"
	"github.com/level11consulting/ocelot/client/buildconfigvalidator"
	"github.com/level11consulting/ocelot/client/newbuildjob"
	"github.com/level11consulting/ocelot/models"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/server/config"
	"github.com/level11consulting/ocelot/storage"
	storage_error "github.com/level11consulting/ocelot/storage/error"
	"github.com/shankj3/go-til/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metrics "github.com/level11consulting/ocelot/server/metrics/admin"
	"github.com/level11consulting/ocelot/client/runtime"
	"github.com/shankj3/go-til/deserialize"
	"github.com/shankj3/go-til/nsqpb"
	"github.com/level11consulting/ocelot/build/vcshandler"
)

type BuildInterface interface {
	BuildRuntime(context.Context, *pb.BuildQuery) (*pb.Builds, error)
	BuildRepoAndHash(*pb.BuildReq, pb.GuideOcelot_BuildRepoAndHashServer) error
	FindWerker(context.Context, *pb.BuildReq) (*pb.BuildRuntimeInfo, error)
}

type BuildAPI struct {
	BuildInterface
	RemoteConfig   config.CVRemoteConfig
	Storage        storage.OcelotStorage
	Deserializer   *deserialize.Deserializer
	Producer       nsqpb.Producer
	OcyValidator   *buildconfigvalidator.OcelotValidator
	handler        models.VCSHandler
}

// This is a copy from signaler_actions.go.
func (g *BuildAPI) GetHandler(cfg *pb.VCSCreds) (models.VCSHandler, string, error) {
	if g.handler != nil {
		return g.handler, "token", nil
	}
	handler, token, err := vcshandler.GetHandler(cfg)
	if err != nil {
		log.IncludeErrField(err).Error()
		return nil, token, status.Errorf(codes.Internal, "Unable to retrieve the bitbucket client config for %s. \n Error: %s", cfg.AcctName, err.Error())
	}
	return handler, token, nil
}

// This is a copy from signaler_actions.go.
func (g *BuildAPI) GetSignaler() *buildjob.Signaler {
	return buildjob.NewSignaler(g.RemoteConfig, g.Deserializer, g.Producer, g.OcyValidator, g.Storage)
}

func (g *BuildAPI) BuildRuntime(ctx context.Context, bq *pb.BuildQuery) (*pb.Builds, error) {
	start := metrics.StartRequest()
	defer metrics.FinishRequest(start)
	if bq.Hash == "" && bq.BuildId == 0 {
		return nil, status.Error(codes.InvalidArgument, "either hash or build id is required")
	}
	buildRtInfo := make(map[string]*pb.BuildRuntimeInfo)
	var err error

	if len(bq.Hash) > 0 {
		//find matching hashes in consul by git hash
		buildRtInfo, err = runtime.GetBuildRuntime(g.RemoteConfig.GetConsul(), bq.Hash)
		if err != nil {
			if _, ok := err.(*runtime.ErrBuildDone); !ok {
				log.IncludeErrField(err).Error("could not get build runtime")
				return nil, status.Error(codes.Internal, "could not get build runtime, err: "+err.Error())
			} else {
				//we set error back to nil so that we can continue with the rest of the logic here
				err = nil
			}
		}

		//add matching hashes in db if exists and add acctname/repo to ones found in consul
		dbResults, err := g.Storage.RetrieveHashStartsWith(bq.Hash)

		if err != nil {
			return &pb.Builds{
				Builds: buildRtInfo,
			}, storage_error.HandleStorageError(err)
		}

		for _, bild := range dbResults {
			if _, ok := buildRtInfo[bild.Hash]; !ok {
				buildRtInfo[bild.Hash] = &pb.BuildRuntimeInfo{
					Hash: bild.Hash,
					// if a result was found in the database but not in GetBuildRuntime, the build is done
					Done: true,
				}
			}
			buildRtInfo[bild.Hash].AcctName = bild.Account
			buildRtInfo[bild.Hash].RepoName = bild.Repo
		}
	}
	//fixme: this is no longer valid to assume that just because the buildId is passed that the build is done. builds are added to the db from the _start_ of the build.
	//if a valid build id passed, go ask db for entries
	if bq.BuildId > 0 {
		buildSum, err := g.Storage.RetrieveSumByBuildId(bq.BuildId)
		if err != nil {
			return &pb.Builds{
				Builds: buildRtInfo,
			}, storage_error.HandleStorageError(err)
		}

		buildRtInfo[buildSum.Hash] = &pb.BuildRuntimeInfo{
			Hash:     buildSum.Hash,
			Done:     true,
			AcctName: buildSum.Account,
			RepoName: buildSum.Repo,
		}
	}

	builds := &pb.Builds{
		Builds: buildRtInfo,
	}
	return builds, err
}

func (g *BuildAPI) BuildRepoAndHash(buildReq *pb.BuildReq, stream pb.GuideOcelot_BuildRepoAndHashServer) error {
	acct, repo, err := stringbuilder.GetAcctRepo(buildReq.AcctRepo)
	if err != nil {
		return status.Error(codes.InvalidArgument, "Bad format of acctRepo, must be account/repo")
	}
	metrics.TriggeredBuilds.WithLabelValues(acct, repo).Inc()

	if buildReq == nil || len(buildReq.AcctRepo) == 0 {
		return status.Error(codes.InvalidArgument, "please pass a valid account/repo_name and hash")
	}

	// get credentials and appropriate VCS handler for the build request's account / repository
	SendStream(stream, "Searching for VCS creds belonging to %s...", buildReq.AcctRepo)
	cfg, err := config.GetVcsCreds(g.Storage, buildReq.AcctRepo, g.RemoteConfig, buildReq.VcsType)
	if err != nil {
		log.IncludeErrField(err).Error()
		switch err.(type) {
		case *stringbuilder.FormatError:
			return status.Error(codes.InvalidArgument, "Format error: "+err.Error())
		case *storage.ErrMultipleVCSTypes:
			return status.Error(codes.InvalidArgument, "There are multiple vcs types for that account. You must include the VcsType field to be able to retrieve credentials for this build. Original error: "+err.Error())
		default:
			return status.Error(codes.Internal, "Could not retrieve vcs creds: "+err.Error())
		}
	}
	SendStream(stream, "Successfully found VCS credentials belonging to %s %s", buildReq.AcctRepo, models.CHECKMARK)
	SendStream(stream, "Validating VCS Credentials...")
	handler, token, grpcErr := g.GetHandler(cfg)
	if grpcErr != nil {
		return grpcErr
	}
	SendStream(stream, "Successfully used VCS Credentials to obtain a token %s", models.CHECKMARK)
	// see if this request's hash has already been built before. if it has, then that means that we can validate the acct/repo in the db against the buildreq one.
	// it also means we can do some partial hash matching, as well as selecting the branch that is associated with the commit if it isn't passed in as request param
	var hashPreviouslyBuilt bool
	var buildSum *pb.BuildSummary
	if buildReq.Hash != "" {
		buildSum, err = g.Storage.RetrieveLatestSum(buildReq.Hash)
		if err != nil {
			if _, ok := err.(*storage.ErrNotFound); !ok {
				log.IncludeErrField(err).Error("could not retrieve latest build summary")
				return status.Error(codes.Internal, fmt.Sprintf("Unable to connect to the database, therefore this operation is not available at this time."))
			}
			SendStream(stream, "There are no previous builds starting with hash %s...", buildReq.Hash)
		}

		hashPreviouslyBuilt = err == nil
	}
	// validate that hte request acct/repo is the same as an entry in the db. if this happens, we want to know about it.
	if hashPreviouslyBuilt && (buildSum.Repo != repo || buildSum.Account != acct) {
		mismatchErr := errors.New(fmt.Sprintf("The account/repo passed (%s) doesn't match with the account/repo (%s) associated with build #%v", buildReq.AcctRepo, buildSum.Account+"/"+buildSum.Repo, buildSum.BuildId))
		log.IncludeErrField(mismatchErr).Error()
		return status.Error(codes.InvalidArgument, mismatchErr.Error())
	}
	var buildBranch, buildHash string
	switch {
	//	do the lookup of latest commit to get full hash
	case buildReq.Hash == "":
		if buildReq.Branch == "" {
			return status.Error(codes.InvalidArgument, "If not sending a hash, then a lookup will be requested off of the Account/Repo and Branch to find the latest commit. Therefore, acctRepo and branch are required fields")
		}
		hist, err := handler.GetBranchLastCommitData(buildReq.AcctRepo, buildReq.Branch)
		if err != nil {
			if _, ok := err.(*models.BranchNotFound); !ok {
				return status.Error(codes.Unavailable, "Unable to retrieve last commit data from bitbucket handler, error from api is: "+err.Error())
			} else {
				return status.Error(codes.InvalidArgument, fmt.Sprintf("Branch %s was not found for repository %s", buildReq.Branch, buildReq.AcctRepo))
			}
		}
		buildBranch = buildReq.Branch
		buildHash = hist.Hash
		SendStream(stream, "Building branch %s with the latest commit in VCS, which is %s", buildBranch, buildHash)
	// user passed hash and branch, if its been built before use the old build to get the full hash, set the request branch / hash
	case buildReq.Hash != "" && buildReq.Branch != "":
		if hashPreviouslyBuilt {
			buildHash = buildSum.Hash
		} else {
			buildHash = buildReq.Hash
		}
		buildBranch = buildReq.Branch
		SendStream(stream, "Building with given hash %s and branch %s", buildHash, buildBranch)
	// use previously looked up build of this hash to get branch info for build
	case buildReq.Hash != "" && buildReq.Branch == "":
		if !hashPreviouslyBuilt {
			noBranchErr := errors.New("Branch is a required field if a previous build starting with the specified hash cannot be found. Please pass the branch flag and try again!")
			log.IncludeErrField(noBranchErr).Error("branch len is 0")
			return status.Error(codes.InvalidArgument, noBranchErr.Error())
		}
		SendStream(stream, "No branch was passed, using `%s` from build #%v instead...", buildSum.Branch, buildSum.BuildId)
		buildHash = buildSum.Hash
		buildBranch = buildSum.Branch
		SendStream(stream, "Found a previous build starting with hash %s, now building branch %s %s", buildReq.Hash, buildBranch, models.CHECKMARK)
	}
	// get build config to do build validation, that this branch is appropriate,
	SendStream(stream, "Retrieving ocelot.yml for %s...", buildReq.AcctRepo)
	buildConf, err := download.GetConfig(buildReq.AcctRepo, buildHash, g.Deserializer, handler)
	if err != nil {
		log.IncludeErrField(err).Error("couldn't get bb config")
		if err.Error() == "could not find raw data at url" {
			err = status.Error(codes.NotFound, fmt.Sprintf("ocelot.yml not found at commit %s for Acct/Repo %s", buildHash, buildReq.AcctRepo))
		} else {
			err = status.Error(codes.InvalidArgument, "Could not get bitbucket ocelot.yml. Error: "+err.Error())
		}
		return err
	}
	SendStream(stream, "Successfully retrieved ocelot.yml for %s %s", buildReq.AcctRepo, models.CHECKMARK)
	SendStream(stream, "Validating and queuing build data for %s...", buildReq.AcctRepo)

	task := buildjob.BuildInitialWerkerTask(buildConf, buildHash, token, buildBranch, buildReq.AcctRepo, pb.SignaledBy_REQUESTED, nil, handler.GetVcsType())
	task.ChangesetData, err = newbuildjob.GenerateNoPreviousHeadChangeset(handler, buildReq.AcctRepo, buildBranch, buildHash)
	if err != nil {
		log.IncludeErrField(err).Error("unable to generate previous head changeset, changeset data will only include branch")
		task.ChangesetData = &pb.ChangesetData{Branch: buildBranch}
		SendStream(stream, "Unable to retrieve files changed for this commit, triggers for stages will only be off of branch and not commit message or files changed.")
	}
	if err = g.GetSignaler().CheckViableThenQueueAndStore(task, buildReq.Force, nil); err != nil {
		if _, ok := err.(*buildconfigvalidator.NotViable); ok {
			log.Log().Info("not queuing because i'm not supposed to, explanation: " + err.Error())
			return status.Error(codes.InvalidArgument, "This failed build queue validation and therefore will not be built. Use Force if you want to override. Error is: "+err.Error())
		}
		log.IncludeErrField(err).Error("couldn't add to build queue or store in db")
		return status.Error(codes.InvalidArgument, "Couldn't add to build queue or store in DB, err: "+err.Error())
	}
	SendStream(stream, "Build started for %s belonging to %s %s", buildHash, buildReq.AcctRepo, models.CHECKMARK)
	return nil
}

func (g *BuildAPI) FindWerker(ctx context.Context, br *pb.BuildReq) (*pb.BuildRuntimeInfo, error) {
	start := metrics.StartRequest()
	defer metrics.FinishRequest(start)
	if len(br.Hash) > 0 {
		//find matching hashes in consul by git hash
		buildRtInfo, err := runtime.GetBuildRuntime(g.RemoteConfig.GetConsul(), br.Hash)
		if err != nil {
			if _, ok := err.(*runtime.ErrBuildDone); !ok {
				return nil, status.Errorf(codes.Internal, "could not get build runtime, err: %s", err.Error())
			}
			return nil, status.Error(codes.InvalidArgument, "werker not found for request as it has already finished ")
		}

		if len(buildRtInfo) == 0 || len(buildRtInfo) > 1 {
			return nil, status.Error(codes.InvalidArgument, "ONE and ONE ONLY match should be found for your hash")
		}

		for _, v := range buildRtInfo {
			return v, nil
		}
	} else {
		return nil, status.Error(codes.InvalidArgument, "Please pass a hash")
	}
	return nil, nil
}

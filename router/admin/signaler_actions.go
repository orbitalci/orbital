package admin

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shankj3/ocelot/build"
	"github.com/shankj3/ocelot/common/remote"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/shankj3/go-til/log"
	signal "github.com/shankj3/ocelot/build_signaler"
	"github.com/shankj3/ocelot/common"
	cred "github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"
)

var (
	triggeredBuilds = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "admin_triggered_builds",
		Help: "builds triggered by a call to admin",
	}, []string{"account", "repository"})
)

func init() {
	prometheus.MustRegister(triggeredBuilds)
}

func (g *guideOcelotServer) BuildRepoAndHash(buildReq *pb.BuildReq, stream pb.GuideOcelot_BuildRepoAndHashServer) error {
	acct, repo, err := common.GetAcctRepo(buildReq.AcctRepo)
	if err != nil {
		return status.Error(codes.InvalidArgument, "Bad format of acctRepo, must be account/repo")
	}
	triggeredBuilds.WithLabelValues(acct, repo).Inc()

	if buildReq == nil || len(buildReq.AcctRepo) == 0 {
		return status.Error(codes.InvalidArgument, "please pass a valid account/repo_name and hash")
	}

	// get credentials and appropriate VCS handler for the build request's account / repository
	stream.Send(RespWrap(fmt.Sprintf("Searching for VCS creds belonging to %s...", buildReq.AcctRepo)))
	cfg, err := cred.GetVcsCreds(g.Storage, buildReq.AcctRepo, g.RemoteConfig)
	if err != nil {
		log.IncludeErrField(err).Error()
		if _, ok := err.(*common.FormatError); ok {
			return status.Error(codes.InvalidArgument, "Format error: "+err.Error())
		}
		return status.Error(codes.Internal, "Could not retrieve vcs creds: "+err.Error())
	}
	stream.Send(RespWrap(fmt.Sprintf("Successfully found VCS credentials belonging to %s %s", buildReq.AcctRepo, models.CHECKMARK)))
	stream.Send(RespWrap("Validating VCS Credentials..."))
	handler, token, grpcErr := g.getHandler(cfg)
	if grpcErr != nil {
		return grpcErr
	}
	stream.Send(RespWrap(fmt.Sprintf("Successfully used VCS Credentials to obtain a token %s", models.CHECKMARK)))
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
			warnMsg := fmt.Sprintf("There are no previous builds starting with hash %s...", buildReq.Hash)
			stream.Send(RespWrap(warnMsg))
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
		stream.Send(RespWrap(fmt.Sprintf("Building branch %s with the latest commit in VCS, which is %s", buildBranch, buildHash)))
	// user passed hash and branch, if its been built before use the old build to get the full hash, set the request branch / hash
	case buildReq.Hash != "" && buildReq.Branch != "":
		if hashPreviouslyBuilt {
			buildHash = buildSum.Hash
		} else {
			buildHash = buildReq.Hash
		}
		buildBranch = buildReq.Branch
		stream.Send(RespWrap(fmt.Sprintf("Building with given hash %s and branch %s", buildHash, buildBranch)))
	// use previously looked up build of this hash to get branch info for build
	case buildReq.Hash != "" && buildReq.Branch == "":
		if !hashPreviouslyBuilt {
			noBranchErr := errors.New("Branch is a required field if a previous build starting with the specified hash cannot be found. Please pass the branch flag and try again!")
			log.IncludeErrField(noBranchErr).Error("branch len is 0")
			return status.Error(codes.InvalidArgument, noBranchErr.Error())
		}
		stream.Send(RespWrap(fmt.Sprintf("No branch was passed, using `%s` from build #%v instead...", buildSum.Branch, buildSum.BuildId)))
		buildHash = buildSum.Hash
		buildBranch = buildSum.Branch
		stream.Send(RespWrap(fmt.Sprintf("Found a previous build starting with hash %s, now building branch %s %s", buildReq.Hash, buildBranch, models.CHECKMARK)))
	}
	// get build config to do build validation, that this branch is appropriate,
	stream.Send(RespWrap(fmt.Sprintf("Retrieving ocelot.yml for %s...", buildReq.AcctRepo)))
	buildConf, err := signal.GetConfig(buildReq.AcctRepo, buildHash, g.Deserializer, handler)
	if err != nil {
		log.IncludeErrField(err).Error("couldn't get bb config")
		if err.Error() == "could not find raw data at url" {
			err = status.Error(codes.NotFound, fmt.Sprintf("ocelot.yml not found at commit %s for Acct/Repo %s", buildHash, buildReq.AcctRepo))
		} else {
			err = status.Error(codes.InvalidArgument, "Could not get bitbucket ocelot.yml. Error: "+err.Error())
		}
		return err
	}
	stream.Send(RespWrap(fmt.Sprintf("Successfully retrieved ocelot.yml for %s %s", buildReq.AcctRepo, models.CHECKMARK)))
	stream.Send(RespWrap(fmt.Sprintf("Validating and queuing build data for %s...", buildReq.AcctRepo)))
	// i was trying to make this work, but it ends up being really complicated considering that we're dealing with a DAG and (at least) bitbucket's api is not robust in this respect..
	// 	might be worth revisiting, idk, but its not worth it right now.
	//
	//
	// Attempt to get a list of commits from the requested hash back to the last hash that was built. If anything goes wrong here, that's fine we are just going to send an error over the stream then build it anyway.
	//var commits []*pb.Commit
	//sums, err := g.Storage.RetrieveLastFewSums(acct, repo, 1)
	//if err != nil {
	//	log.IncludeErrField(err).Error("could not retrieve last build for acct/repo " + buildReq.AcctRepo)
	//	stream.Send(RespWrap(fmt.Sprintf("Could not retrive last build for acct/repo %s, therefore cannot search commit history for skip commit messages. Just FYI.", buildReq.AcctRepo)))
	//} else {
	//	if len(sums) != 1 {
	//		log.Log().Errorf("length of retrieved summaries for build request %s %s is %d.. wtf?", buildReq.AcctRepo, buildReq.Hash, len(sums))
	//		stream.Send(RespWrap(fmt.Sprintf("Error retrieving last build for acct/repo %s, therefore cannot search commit history for skip commit messages. Just FYI.", buildReq.AcctRepo)))
	//	} else {
	//
	//		commits, err = handler.GetCommitLog(buildReq.AcctRepo, branch, sums[0].Hash)
	//	}
	//}
	task := signal.BuildInitialWerkerTask(buildConf, buildHash, token, buildBranch, buildReq.AcctRepo, pb.SignaledBy_REQUESTED, nil)
	task.ChangesetData, err = signal.GenerateNoPreviousHeadChangeset(handler, buildReq.AcctRepo, buildBranch, buildHash)
	if err != nil {
		log.IncludeErrField(err).Error("unable to generate previous head changeset, changeset data will only include branch")
		task.ChangesetData = &pb.ChangesetData{Branch: buildBranch}
		stream.Send(RespWrap(fmt.Sprintf("Unable to retrieve files changed for this commit, triggers for stages will only be off of branch and not commit message or files changed.")))
	}
	if err = g.getSignaler().CheckViableThenQueueAndStore(task, buildReq.Force, nil); err != nil {
		if _, ok := err.(*build.NotViable); ok {
			log.Log().Info("not queuing because i'm not supposed to, explanation: " + err.Error())
			return status.Error(codes.InvalidArgument, "This failed build queue validation and therefore will not be built. Use Force if you want to override. Error is: "+err.Error())
		}
		log.IncludeErrField(err).Error("couldn't add to build queue or store in db")
		return status.Error(codes.InvalidArgument, "Couldn't add to build queue or store in DB, err: "+err.Error())
	}
	stream.Send(RespWrap(fmt.Sprintf("Build started for %s belonging to %s %s", buildHash, buildReq.AcctRepo, models.CHECKMARK)))
	return nil
}

// getHandler returns a grpc status.Error
func (g *guideOcelotServer) getHandler(cfg *pb.VCSCreds) (models.VCSHandler, string, error) {
	if g.handler != nil {
		return g.handler, "token", nil
	}
	handler, token, err := remote.GetHandler(cfg)
	if err != nil {
		log.IncludeErrField(err).Error()
		return nil, token, status.Errorf(codes.Internal, "Unable to retrieve the bitbucket client config for %s. \n Error: %s", cfg.AcctName, err.Error())
	}
	return handler, token, nil
}

func (g *guideOcelotServer) getSignaler() *signal.Signaler {
	return signal.NewSignaler(g.RemoteConfig, g.Deserializer, g.Producer, g.OcyValidator, g.Storage)
}

func (g *guideOcelotServer) WatchRepo(ctx context.Context, repoAcct *pb.RepoAccount) (*empty.Empty, error) {
	if repoAcct.Repo == "" || repoAcct.Account == "" {
		return nil, status.Error(codes.InvalidArgument, "repo and account are required fields")
	}
	cfg, err := cred.GetVcsCreds(g.Storage, repoAcct.Account+"/"+repoAcct.Repo, g.RemoteConfig)
	if err != nil {
		log.IncludeErrField(err).Error()
		if _, ok := err.(*common.FormatError); ok {
			return nil, status.Error(codes.InvalidArgument, "Format error: "+err.Error())
		}
		return nil, status.Error(codes.Internal, "Could not retrieve vcs creds: "+err.Error())
	}
	handler, _, grpcErr := g.getHandler(cfg)
	if grpcErr != nil {
		return nil, grpcErr
	}
	repoLinks, err := handler.GetRepoLinks(fmt.Sprintf("%s/%s", repoAcct.Account, repoAcct.Repo))
	if err != nil {
		return &empty.Empty{}, status.Errorf(codes.Unavailable, "could not get repository detail at %s/%s", repoAcct.Account, repoAcct.Repo)
	}
	err = handler.CreateWebhook(repoLinks.Hooks)

	if err != nil {
		return &empty.Empty{}, status.Error(codes.Unavailable, errors.WithMessage(err, "Unable to create webhook").Error())
	}
	return &empty.Empty{}, nil
}

func (g *guideOcelotServer) PollRepo(ctx context.Context, poll *pb.PollRequest) (*empty.Empty, error) {
	if poll.Account == "" || poll.Repo == "" || poll.Cron == "" || poll.Branches == "" {
		return nil, status.Error(codes.InvalidArgument, "account, repo, cron, and branches are required fields")
	}
	log.Log().Info("recieved poll request for ", poll.Account, poll.Repo, poll.Cron)
	empti := &empty.Empty{}
	exists, err := g.Storage.PollExists(poll.Account, poll.Repo)
	if err != nil {
		return empti, status.Error(codes.Unavailable, "unable to retrieve poll table from storage. err: "+err.Error())
	}
	if exists == true {
		log.Log().Info("updating poll in db")
		if err = g.Storage.UpdatePoll(poll.Account, poll.Repo, poll.Cron, poll.Branches); err != nil {
			msg := "unable to update poll in storage"
			log.IncludeErrField(err).Error(msg)
			return empti, status.Error(codes.Unavailable, msg+": "+err.Error())
		}
	} else {
		log.Log().Info("inserting poll in db")
		if err = g.Storage.InsertPoll(poll.Account, poll.Repo, poll.Cron, poll.Branches); err != nil {
			msg := "unable to insert poll into storage"
			log.IncludeErrField(err).Error(msg)
			return empti, status.Error(codes.Unavailable, msg+": "+err.Error())
		}
	}
	log.Log().WithField("account", poll.Account).WithField("repo", poll.Repo).Info("successfully added/updated poll in storage")
	err = g.Producer.WriteProto(poll, "poll_please")
	if err != nil {
		log.IncludeErrField(err).Error("couldn't write to queue producer at poll_please")
		return empti, status.Error(codes.Unavailable, err.Error())
	}
	return empti, nil
}

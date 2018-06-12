package admin

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
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

func (g *guideOcelotServer) BuildRepoAndHash(buildReq *pb.BuildReq, stream pb.GuideOcelot_BuildRepoAndHashServer) error {
	acct, repo, err := common.GetAcctRepo(buildReq.AcctRepo)
	if err != nil {
		return status.Error(codes.InvalidArgument, "Bad format of acctRepo, must be account/repo")
	}
	log.Log().Info(buildReq)
	if buildReq == nil || len(buildReq.AcctRepo) == 0 || len(buildReq.Hash) == 0 {
		return status.Error(codes.InvalidArgument, "please pass a valid account/repo_name and hash")
	}

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
	handler, token, err := remote.GetHandler(cfg)
	if err != nil {
		log.IncludeErrField(err).Error()
		return status.Error(codes.Internal, fmt.Sprintf("Unable to retrieve the bitbucket client config for %s. \n Error: %s", buildReq.AcctRepo, err.Error()))
	}
	stream.Send(RespWrap(fmt.Sprintf("Successfully used VCS Credentials to obtain a token %s", models.CHECKMARK)))

	var branch string
	var fullHash string
	stream.Send(RespWrap(fmt.Sprintf("Looking in previous builds for %s...", buildReq.Hash)))
	buildSum, err := g.Storage.RetrieveLatestSum(buildReq.Hash)
	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); !ok {
			log.IncludeErrField(err).Error("could not retrieve latest build summary")
			return status.Error(codes.Internal, fmt.Sprintf("Unable to connect to the database, therefore this operation is not available at this time."))
		}
		//at this point error must be because we couldn't find hash starting with query
		warnMsg := fmt.Sprintf("There are no previous builds starting with hash %s...", buildReq.Hash)
		log.IncludeErrField(err).Warning(warnMsg)
		stream.Send(RespWrap(warnMsg))

		if len(buildReq.Branch) == 0 {
			noBranchErr := errors.New("Branch is a required field if a previous build starting with the specified hash cannot be found. Please pass the branch flag and try again!")
			log.IncludeErrField(noBranchErr).Error("branch len is 0")
			return status.Error(codes.InvalidArgument, noBranchErr.Error())
		}

		fullHash = buildReq.Hash
		branch = buildReq.Branch
	} else {
		if buildSum.Repo != repo || buildSum.Account != acct {
			mismatchErr := errors.New(fmt.Sprintf("The account/repo passed (%s) doesn't match with the account/repo (%s) associated with build #%v", buildReq.AcctRepo, buildSum.Account+"/"+buildSum.Repo, buildSum.BuildId))
			log.IncludeErrField(mismatchErr).Error()
			return status.Error(codes.InvalidArgument, mismatchErr.Error())
		}

		if len(buildReq.Branch) == 0 {
			stream.Send(RespWrap(fmt.Sprintf("No branch was passed, using `%s` from build #%v instead...", buildSum.Branch, buildSum.BuildId)))
			branch = buildSum.Branch
		} else {
			branch = buildReq.Branch
		}

		fullHash = buildSum.Hash
		stream.Send(RespWrap(fmt.Sprintf("Found a previous build starting with hash %s, now building branch %s %s", buildReq.Hash, branch, models.CHECKMARK)))
	}

	stream.Send(RespWrap(fmt.Sprintf("Retrieving ocelot.yml for %s...", buildReq.AcctRepo)))
	buildConf, err := signal.GetConfig(buildReq.AcctRepo, fullHash, g.Deserializer, handler)
	if err != nil {
		log.IncludeErrField(err).Error("couldn't get bb config")
		if err.Error() == "could not find raw data at url" {
			err = status.Error(codes.NotFound, fmt.Sprintf("File not found at commit %s for Acct/Repo %s", fullHash, buildReq.AcctRepo))
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
	task := signal.BuildInitialWerkerTask(buildConf, buildReq.Hash, token, branch, buildReq.AcctRepo, pb.SignaledBy_REQUESTED, nil)
	if err = g.getSignaler().CheckViableThenQueueAndStore(task, buildReq.Force, nil); err != nil {
		if _, ok := err.(*build.NotViable); ok {
			log.Log().Info("not queuing because i'm not supposed to, explanation: " + err.Error())
			return status.Error(codes.InvalidArgument, "This failed build queue validation and therefore will not be built. Use Force if you want to override. Error is: " + err.Error())
		}
		log.IncludeErrField(err).Error("couldn't add to build queue or store in db")
		return status.Error(codes.InvalidArgument, "Couldn't add to build queue or store in DB, err: "+err.Error())
	}
	stream.Send(RespWrap(fmt.Sprintf("Build started for %s belonging to %s %s", fullHash, buildReq.AcctRepo, models.CHECKMARK)))
	return nil
}

func (g *guideOcelotServer) getSignaler() *signal.Signaler {
	return signal.NewSignaler(g.RemoteConfig, g.Deserializer, g.Producer, g.OcyValidator, g.Storage)
}

func (g *guideOcelotServer) WatchRepo(ctx context.Context, repoAcct *pb.RepoAccount) (*empty.Empty, error) {
	if repoAcct.Repo == "" || repoAcct.Account == "" {
		return nil, status.Error(codes.InvalidArgument, "repo and account are required fields")
	}
	cfg, err := cred.GetVcsCreds(g.Storage, repoAcct.Account + "/" + repoAcct.Repo, g.RemoteConfig)
	if err != nil {
		log.IncludeErrField(err).Error()
		if _, ok := err.(*common.FormatError); ok {
			return nil, status.Error(codes.InvalidArgument, "Format error: "+err.Error())
		}
		return nil, status.Error(codes.Internal, "Could not retrieve vcs creds: "+err.Error())
	}
	handler, _, err := remote.GetHandler(cfg)
	if err != nil {
		log.IncludeErrField(err).Error()
		return nil, status.Error(codes.Internal, "Unable to retrieve the bitbucket client config for %s. \n Error: %s")
	}
	repoDetail, err := handler.GetRepoDetail(fmt.Sprintf("%s/%s", repoAcct.Account, repoAcct.Repo))
	if repoDetail.Type == "error" || err != nil {
		return &empty.Empty{}, status.Errorf(codes.Unavailable, "could not get repository detail at %s/%s", repoAcct.Account, repoAcct.Repo)
	}

	webhookURL := repoDetail.GetLinks().GetHooks().GetHref()
	err = handler.CreateWebhook(webhookURL)

	if err != nil {
		return &empty.Empty{}, status.Error(codes.Unavailable, err.Error())
	}
	return &empty.Empty{}, nil
}

func (g *guideOcelotServer) PollRepo(ctx context.Context, poll *pb.PollRequest) (*empty.Empty, error) {
	if poll.Account == "" || poll.Repo == "" || poll.Cron == "" || poll.Branches == "" {
		return nil, status.Error(codes.InvalidArgument, "account, repo, cron, and branches are required fields")
	}
	log.Log().Info("recieved poll request for ", poll.Account, poll.Repo, poll.Cron)
	empti := &empty.Empty{}
	if poll.Repo == "" || poll.Account == "" || poll.Branches == "" || poll.Cron == "" {
		return empti, status.Error(codes.InvalidArgument, "account, poll, repo, and cron are all required fields")
	}
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

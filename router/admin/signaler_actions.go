package admin

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/go-til/net"
	signal "bitbucket.org/level11consulting/ocelot/build_signaler"
	"bitbucket.org/level11consulting/ocelot/common"
	cred "bitbucket.org/level11consulting/ocelot/common/credentials"
	bbh "bitbucket.org/level11consulting/ocelot/common/remote/bitbucket"
	"bitbucket.org/level11consulting/ocelot/models"
	"bitbucket.org/level11consulting/ocelot/models/pb"
	"bitbucket.org/level11consulting/ocelot/storage"

)


func (g *guideOcelotServer) BuildRepoAndHash(buildReq *pb.BuildReq, stream pb.GuideOcelot_BuildRepoAndHashServer) error {
	log.Log().Info(buildReq)
	if buildReq == nil || len(buildReq.AcctRepo) == 0 || len(buildReq.Hash) == 0 {
		return status.Error(codes.InvalidArgument, "please pass a valid account/repo_name and hash")
	}

	stream.Send(RespWrap(fmt.Sprintf("Searching for VCS creds belonging to %s...", buildReq.AcctRepo)))
	cfg, err := cred.GetVcsCreds(g.Storage, buildReq.AcctRepo, g.RemoteConfig)
	if err != nil {
		log.IncludeErrField(err).Error()
		if _, ok := err.(*common.FormatError); ok {
			return status.Error(codes.InvalidArgument, "Format error: " + err.Error())
		}
		return status.Error(codes.Internal, "Could not retrieve vcs creds: " + err.Error())
	}
	stream.Send(RespWrap(fmt.Sprintf("Successfully found VCS credentials belonging to %s %s", buildReq.AcctRepo, models.CHECKMARK)))
	stream.Send(RespWrap("Validating VCS Credentials..."))
	bbHandler, token, err := bbh.GetBitbucketClient(cfg)
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
		acct, repo, err := common.GetAcctRepo(buildReq.AcctRepo)
		if err != nil {
			return status.Error(codes.InvalidArgument, "Bad format of acctRepo, must be account/repo")
		}
		if buildSum.Repo != repo || buildSum.Account != acct {
			mismatchErr := errors.New(fmt.Sprintf("The account/repo passed (%s) doesn't match with the account/repo (%s) associated with build #%v", buildReq.AcctRepo, buildSum.Account + "/" + buildSum.Repo, buildSum.BuildId))
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
	buildConf, _, err := signal.GetBBConfig(g.RemoteConfig, g.Storage, buildReq.AcctRepo, fullHash, g.Deserializer, bbHandler)
	if err != nil {
		log.IncludeErrField(err).Error("couldn't get bb config")
		if err.Error() == "could not find raw data at url" {
			err = status.Error(codes.NotFound, fmt.Sprintf("File not found at commit %s for Acct/Repo %s", fullHash, buildReq.AcctRepo))
		} else {
			err = status.Error(codes.InvalidArgument, "Could not get bitbucket ocelot.yml. Error: " + err.Error())
		}
		return err
	}
	stream.Send(RespWrap(fmt.Sprintf("Successfully retrieved ocelot.yml for %s %s", buildReq.AcctRepo, models.CHECKMARK)))
	stream.Send(RespWrap(fmt.Sprintf("Storing build data for %s...", buildReq.AcctRepo)))
	if err = signal.QueueAndStore(fullHash, branch, buildReq.AcctRepo, token, g.RemoteConfig, buildConf, g.OcyValidator, g.Producer, g.Storage); err != nil {
		log.IncludeErrField(err).Error("couldn't add to build queue or store in db")
		return status.Error(codes.InvalidArgument, "Couldn't add to build queue or store in DB, err: " + err.Error())
	}
	stream.Send(RespWrap(fmt.Sprintf("Build started for %s belonging to %s %s", fullHash, buildReq.AcctRepo, models.CHECKMARK)))
	return nil
}



func (g *guideOcelotServer) WatchRepo(ctx context.Context, repoAcct *pb.RepoAccount) (*empty.Empty, error) {
	var vcs *pb.VCSCreds
	bb := pb.SubCredType_BITBUCKET
	identifier, err := pb.CreateVCSIdentifier(bb, repoAcct.Account)
	if err != nil {
		return &empty.Empty{}, status.Error(codes.Internal, "couldn't create identifier")
	}
	bbCreds, err := g.RemoteConfig.GetCred(g.Storage, bb, identifier, repoAcct.Account, true)
	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); ok {
			return &empty.Empty{}, status.Error(codes.NotFound, "credentials not found")
		}
		return &empty.Empty{}, status.Error(codes.Internal, "could not get bitbucket creds")
	}
	vcs = bbCreds.(*pb.VCSCreds)
	bbClient := &net.OAuthClient{}
	bbClient.Setup(vcs)


	bbHandler := bbh.GetBitbucketHandler(vcs, bbClient)
	repoDetail, err := bbHandler.GetRepoDetail(fmt.Sprintf("%s/%s", repoAcct.Account, repoAcct.Repo))
	if repoDetail.Type == "error" || err != nil {
		return &empty.Empty{}, status.Errorf(codes.Unavailable, "could not get repository detail at %s/%s", repoAcct.Account, repoAcct.Repo)
	}

	webhookURL := repoDetail.GetLinks().GetHooks().GetHref()
	err = bbHandler.CreateWebhook(webhookURL)

	if err != nil {
		return &empty.Empty{}, status.Error(codes.Unavailable, err.Error())
	}
	return &empty.Empty{}, nil
}


func (g *guideOcelotServer) PollRepo(ctx context.Context, poll *pb.PollRequest) (*empty.Empty, error) {
	log.Log().Info("recieved poll request for ", poll.Account, poll.Repo, poll.Cron)
	empti := &empty.Empty{}
	if poll.Repo == "" || poll.Account == "" || poll.Branches == "" || poll.Cron == "" {
		return empti, status.Error(codes.InvalidArgument, "account, poll, repo, and cron are all required fields")
	}
	exists, err := g.Storage.PollExists(poll.Account, poll.Repo)
	if err != nil {
		return empti, status.Error(codes.Unavailable, "unable to retrieve poll table from storage. err: " + err.Error())
	}
	if exists == true {
		log.Log().Info("updating poll in db")
		if err = g.Storage.UpdatePoll(poll.Account, poll.Repo, poll.Cron, poll.Branches); err != nil {
			msg := "unable to update poll in storage"
			log.IncludeErrField(err).Error(msg)
			return empti, status.Error(codes.Unavailable, msg + ": " + err.Error())
		}
	} else {
		log.Log().Info("inserting poll in db")
		if err = g.Storage.InsertPoll(poll.Account, poll.Repo, poll.Cron, poll.Branches); err != nil {
			msg := "unable to insert poll into storage"
			log.IncludeErrField(err).Error(msg)
			return empti, status.Error(codes.Unavailable, msg + ": " + err.Error())
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

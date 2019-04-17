package admin

import (
	"context"
	"github.com/pkg/errors"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	stringbuilder "github.com/level11consulting/ocelot/build/helpers/stringbuilder/accountrepo"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/server/config"
	"github.com/shankj3/go-til/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/level11consulting/ocelot/storage"
)

type RepoInterface interface {
	PollSchedule
	WatchRepo(context.Context, *pb.RepoAccount) (*empty.Empty, error)
	GetTrackedRepos(context.Context, *empty.Empty) (*pb.AcctRepos, error)
}

func (g *OcelotServerAPI) WatchRepo(ctx context.Context, repoAcct *pb.RepoAccount) (*empty.Empty, error) {
	if repoAcct.Repo == "" || repoAcct.Account == "" || repoAcct.Type == pb.SubCredType_NIL_SCT {
		return nil, status.Error(codes.InvalidArgument, "repo, account, and type are required fields")
	}
	// check to make sure url for webhook exists before trying anything fancy
	if g.DeprecatedHandler.hhBaseUrl == "" {
		return &empty.Empty{}, status.Error(codes.Unimplemented, "Admin is not configured with a hookhandler callback url to register webhooks with. Contact your administrator to run the ocelot admin service with the flag -hookhandler-url-base set to a url that can be accessed via a webhook for VCS push/pullrequest events.")
	}
	cfg, err := config.GetVcsCreds(g.DeprecatedHandler.Storage, repoAcct.Account+"/"+repoAcct.Repo, g.DeprecatedHandler.RemoteConfig, repoAcct.Type)
	if err != nil {
		log.IncludeErrField(err).Error()
		if _, ok := err.(*stringbuilder.FormatError); ok {
			return nil, status.Error(codes.InvalidArgument, "Format error: "+err.Error())
		}
		return nil, status.Error(codes.Internal, "Could not retrieve vcs creds: "+err.Error())
	}
	handler, _, grpcErr := g.GetHandler(cfg)
	if grpcErr != nil {
		return nil, grpcErr
	}
	repoLinks, err := handler.GetRepoLinks(fmt.Sprintf("%s/%s", repoAcct.Account, repoAcct.Repo))
	if err != nil {
		return &empty.Empty{}, status.Errorf(codes.Unavailable, "could not get repository detail at %s/%s", repoAcct.Account, repoAcct.Repo)
	}
	handler.SetCallbackURL(g.DeprecatedHandler.hhBaseUrl)
	err = handler.CreateWebhook(repoLinks.Hooks)

	if err != nil {
		return &empty.Empty{}, status.Error(codes.Unavailable, errors.WithMessage(err, "Unable to create webhook").Error())
	}
	return &empty.Empty{}, nil
}

func (g *OcelotServerAPI) GetTrackedRepos(ctx context.Context, empty *empty.Empty) (*pb.AcctRepos, error) {
	repos, err := g.DeprecatedHandler.Storage.GetTrackedRepos()
	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); ok {
			return nil, status.Error(codes.NotFound, "could not find any account/repos in the database")
		}
		return nil, status.Error(codes.FailedPrecondition, "an error occurred getting account/repos from db")
	}
	return repos, nil
}
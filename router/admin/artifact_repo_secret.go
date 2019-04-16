package admin

import (
	"context"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ArtifactRepoSecret interface {
	GetRepoCred(context.Context, *pb.RepoCreds) (*pb.RepoCreds, error)
	RepoCredExists(context.Context, *pb.RepoCreds) (*pb.Exists, error)
	UpdateRepoCreds(context.Context, *pb.RepoCreds) (*empty.Empty, error)
	GetRepoCreds(context.Context, *empty.Empty) (*pb.RepoCredWrapper, error)
	SetRepoCreds(context.Context, *pb.RepoCreds) (*empty.Empty, error)
	DeleteRepoCreds(context.Context, *pb.RepoCreds) (*empty.Empty, error)
}

func (g *guideOcelotServer) GetRepoCreds(ctx context.Context, msg *empty.Empty) (*pb.RepoCredWrapper, error) {
	credWrapper := &pb.RepoCredWrapper{}
	creds, err := g.RemoteConfig.GetCredsByType(g.Storage, pb.CredType_REPO, true)

	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); !ok {
			return credWrapper, status.Error(codes.ResourceExhausted, err.Error())
		}
		return credWrapper, status.Error(codes.NotFound, err.Error())
	}

	for _, v := range creds {
		credWrapper.Repo = append(credWrapper.Repo, v.(*pb.RepoCreds))
	}
	if len(credWrapper.Repo) == 0 {
		return nil, status.Error(codes.NotFound, "no repo creds found")
	}
	return credWrapper, nil
}

func (g *guideOcelotServer) GetRepoCred(ctx context.Context, credentials *pb.RepoCreds) (*pb.RepoCreds, error) {
	creddy, err := g.getAnyCred(credentials)
	if err != nil {
		return nil, err
	}
	repo, ok := creddy.(*pb.RepoCreds)
	if !ok {
		return nil, status.Error(codes.Internal, "Unable to cast as Repo Creds")
	}
	return repo, nil
}

func (g *guideOcelotServer) SetRepoCreds(ctx context.Context, creds *pb.RepoCreds) (*empty.Empty, error) {
	if creds.SubType.Parent() != pb.CredType_REPO {
		return nil, status.Error(codes.InvalidArgument, "Subtype must be of repo type: "+strings.Join(pb.CredType_REPO.SubtypesString(), " | "))
	}
	err := g.RepoValidator.ValidateConfig(creds)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed repo creds validation! error: %s", err.Error())
	}
	err = SetupRCCCredentials(g.RemoteConfig, g.Storage, creds)
	if _, ok := err.(*pb.ValidationErr); ok {
		return &empty.Empty{}, status.Error(codes.InvalidArgument, "Repo Creds failed validation. Errors are: "+err.Error())
	}
	if err != nil {
		return &empty.Empty{}, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}

func (g *guideOcelotServer) UpdateRepoCreds(ctx context.Context, creds *pb.RepoCreds) (*empty.Empty, error) {
	return g.updateAnyCred(ctx, creds)
}

func (g *guideOcelotServer) RepoCredExists(ctx context.Context, creds *pb.RepoCreds) (*pb.Exists, error) {
	return g.checkAnyCredExists(ctx, creds)
}

func (g *guideOcelotServer) DeleteRepoCreds(ctx context.Context, creds *pb.RepoCreds) (*empty.Empty, error) {
	return g.deleteAnyCred(ctx, creds, pb.CredType_REPO)
}
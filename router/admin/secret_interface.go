package admin

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SecretInterface interface {
	KubernetesSecret
	NotifierSecret
	SshSecret
	VcsSecret
	AppleDevSecret
	GenericSecret
	ArtifactRepoSecret
	GetAllCreds(context.Context, *empty.Empty) (*pb.AllCredsWrapper, error)
}

func (g *guideOcelotServer) GetAllCreds(ctx context.Context, msg *empty.Empty) (*pb.AllCredsWrapper, error) {
	allCreds := &pb.AllCredsWrapper{}
	repoCreds, err := g.GetRepoCreds(ctx, msg)
	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); ok {
			return allCreds, status.Error(codes.NotFound, err.Error())
		}
		return allCreds, status.Errorf(codes.Internal, "unable to get repo creds! error: %s", err.Error())
	}
	allCreds.RepoCreds = repoCreds
	adminCreds, err := g.GetVCSCreds(ctx, msg)
	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); ok {
			return allCreds, status.Error(codes.NotFound, err.Error())
		}
		return allCreds, status.Errorf(codes.Internal, "unable to get vcs creds! error: %s", err.Error())
	}
	allCreds.VcsCreds = adminCreds
	return allCreds, nil
}

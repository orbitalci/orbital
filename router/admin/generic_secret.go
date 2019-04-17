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

type GenericSecret interface {
	GetGenericCreds(context.Context, *empty.Empty) (*pb.GenericWrap, error)
	SetGenericCreds(context.Context, *pb.GenericCreds) (*empty.Empty, error)
	UpdateGenericCreds(context.Context, *pb.GenericCreds) (*empty.Empty, error)
	GenericCredExists(context.Context, *pb.GenericCreds) (*pb.Exists, error)
	DeleteGenericCreds(context.Context, *pb.GenericCreds) (*empty.Empty, error)
}

func (g *OcelotServerAPI) GetGenericCreds(ctx context.Context, empty *empty.Empty) (*pb.GenericWrap, error) {
	credz := &pb.GenericWrap{}
	creds, err := g.RemoteConfig.GetCredsByType(g.Storage, pb.CredType_GENERIC, true)
	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); ok {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return credz, status.Errorf(codes.Internal, "unable to get generic creds! error: %s", err.Error())
	}
	for _, v := range creds {
		credz.Creds = append(credz.Creds, v.(*pb.GenericCreds))
	}
	if len(credz.Creds) == 0 {
		return nil, status.Error(codes.NotFound, "no generic credentials found")
	}
	return credz, nil
}

func (g *OcelotServerAPI) SetGenericCreds(ctx context.Context, creds *pb.GenericCreds) (*empty.Empty, error) {
	if creds.SubType.Parent() != pb.CredType_GENERIC {
		return nil, status.Error(codes.InvalidArgument, "Subtype must be of generic type: "+strings.Join(pb.CredType_SSH.SubtypesString(), " | "))
	}
	err := SetupRCCCredentials(g.RemoteConfig, g.Storage, creds)
	if err != nil {
		if _, ok := err.(*pb.ValidationErr); ok {
			return &empty.Empty{}, status.Error(codes.FailedPrecondition, "Generic Creds Upload failed validation. Errors are: "+err.Error())
		}
		return &empty.Empty{}, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}

func (g *OcelotServerAPI) DeleteGenericCreds(ctx context.Context, creds *pb.GenericCreds) (*empty.Empty, error) {
	return g.DeleteAnyCred(ctx, creds, pb.CredType_GENERIC)
}

func (g *OcelotServerAPI) GenericCredExists(ctx context.Context, creds *pb.GenericCreds) (*pb.Exists, error) {
	return g.CheckAnyCredExists(ctx, creds)
}

func (g *OcelotServerAPI) UpdateGenericCreds(ctx context.Context, creds *pb.GenericCreds) (*empty.Empty, error) {
	return g.UpdateAnyCred(ctx, creds)
}

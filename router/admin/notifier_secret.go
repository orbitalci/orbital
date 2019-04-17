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

type NotifierSecret interface {
	SetNotifyCreds(context.Context, *pb.NotifyCreds) (*empty.Empty, error)
	GetNotifyCred(context.Context, *pb.NotifyCreds) (*pb.NotifyCreds, error)
	GetNotifyCreds(context.Context, *empty.Empty) (*pb.NotifyWrap, error)
	UpdateNotifyCreds(context.Context, *pb.NotifyCreds) (*empty.Empty, error)
	NotifyCredExists(context.Context, *pb.NotifyCreds) (*pb.Exists, error)
	DeleteNotifyCreds(context.Context, *pb.NotifyCreds) (*empty.Empty, error)
}

func (g *OcelotServerAPI) SetNotifyCreds(ctx context.Context, creds *pb.NotifyCreds) (*empty.Empty, error) {
	if creds.SubType.Parent() != pb.CredType_NOTIFIER {
		return nil, status.Error(codes.InvalidArgument, "Subtype must be of notifier type: "+strings.Join(pb.CredType_SSH.SubtypesString(), " | "))
	}
	err := SetupRCCCredentials(g.RemoteConfig, g.Storage, creds)
	if err != nil {
		if _, ok := err.(*pb.ValidationErr); ok {
			return &empty.Empty{}, status.Error(codes.FailedPrecondition, "Notify Creds Upload failed validation. Errors are: "+err.Error())
		}
		return &empty.Empty{}, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}

func (g *OcelotServerAPI) NotifyCredExists(ctx context.Context, creds *pb.NotifyCreds) (*pb.Exists, error) {
	return g.CheckAnyCredExists(ctx, creds)
}

func (g *OcelotServerAPI) UpdateNotifyCreds(ctx context.Context, creds *pb.NotifyCreds) (*empty.Empty, error) {
	return g.UpdateAnyCred(ctx, creds)
}

func (g *OcelotServerAPI) GetNotifyCred(ctx context.Context, creds *pb.NotifyCreds) (*pb.NotifyCreds, error) {
	creddy, err := g.GetAnyCred(creds)
	if err != nil {
		return nil, err
	}
	notifier, ok := creddy.(*pb.NotifyCreds)
	if !ok {
		return nil, status.Error(codes.Internal, "Unable to cast as Notifier Creds")
	}
	return notifier, nil
}

func (g *OcelotServerAPI) GetNotifyCreds(ctx context.Context, empty2 *empty.Empty) (*pb.NotifyWrap, error) {
	credWrapper := &pb.NotifyWrap{}
	credz, err := g.RemoteConfig.GetCredsByType(g.Storage, pb.CredType_NOTIFIER, true)
	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); ok {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return credWrapper, status.Errorf(codes.Internal, "unable to get notify creds! error: %s", err.Error())
	}
	for _, v := range credz {
		credWrapper.Creds = append(credWrapper.Creds, v.(*pb.NotifyCreds))
	}
	if len(credWrapper.Creds) == 0 {
		return credWrapper, status.Error(codes.NotFound, "no notifier creds found")
	}
	return credWrapper, nil
}

func (g *OcelotServerAPI) DeleteNotifyCreds(ctx context.Context, creds *pb.NotifyCreds) (*empty.Empty, error) {
	return g.DeleteAnyCred(ctx, creds, pb.CredType_NOTIFIER)
}

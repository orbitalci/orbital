package notifier

import (
	"fmt"
	"context"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/level11consulting/orbitalci/models/pb"
	"github.com/level11consulting/orbitalci/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/level11consulting/orbitalci/secret/legacy"
	"github.com/level11consulting/orbitalci/secret/anycred"
	"github.com/level11consulting/orbitalci/server/config"
)

type NotifierSecret interface {
	SetNotifyCreds(context.Context, *pb.NotifyCreds) (*empty.Empty, error)
	GetNotifyCred(context.Context, *pb.NotifyCreds) (*pb.NotifyCreds, error)
	GetNotifyCreds(context.Context, *empty.Empty) (*pb.NotifyWrap, error)
	UpdateNotifyCreds(context.Context, *pb.NotifyCreds) (*empty.Empty, error)
	NotifyCredExists(context.Context, *pb.NotifyCreds) (*pb.Exists, error)
	DeleteNotifyCreds(context.Context, *pb.NotifyCreds) (*empty.Empty, error)
}

type NotifierSecretAPI struct {
	NotifierSecret
	RemoteConfig   config.CVRemoteConfig
	Storage        storage.OcelotStorage
}

func (g *NotifierSecretAPI) SetNotifyCreds(ctx context.Context, creds *pb.NotifyCreds) (*empty.Empty, error) {
	if creds.SubType.Parent() != pb.CredType_NOTIFIER {
		return nil, status.Error(codes.InvalidArgument, "Subtype must be of notifier type: "+strings.Join(pb.CredType_SSH.SubtypesString(), " | "))
	}
	err := legacy.SetupRCCCredentials(g.RemoteConfig, g.Storage, creds)
	if err != nil {
		if _, ok := err.(*pb.ValidationErr); ok {
			return &empty.Empty{}, status.Error(codes.FailedPrecondition, "Notify Creds Upload failed validation. Errors are: "+err.Error())
		}
		return &empty.Empty{}, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}

func (g *NotifierSecretAPI) NotifyCredExists(ctx context.Context, creds *pb.NotifyCreds) (*pb.Exists, error) {
	anyCredAPI := anycred.AnyCredAPI {
		Storage:        g.Storage,	
		RemoteConfig:   g.RemoteConfig,
	}

	return anyCredAPI.CheckAnyCredExists(ctx, creds)
}

func (g *NotifierSecretAPI) UpdateNotifyCreds(ctx context.Context, creds *pb.NotifyCreds) (*empty.Empty, error) {
	anyCredAPI := anycred.AnyCredAPI {
		Storage:        g.Storage,	
		RemoteConfig:   g.RemoteConfig,
	}

	return anyCredAPI.UpdateAnyCred(ctx, creds)
}

func (g *NotifierSecretAPI) GetNotifyCred(ctx context.Context, creds *pb.NotifyCreds) (*pb.NotifyCreds, error) {
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

func (g *NotifierSecretAPI) GetNotifyCreds(ctx context.Context, empty2 *empty.Empty) (*pb.NotifyWrap, error) {
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

func (g *NotifierSecretAPI) DeleteNotifyCreds(ctx context.Context, creds *pb.NotifyCreds) (*empty.Empty, error) {
	anyCredAPI := anycred.AnyCredAPI {
		Storage:        g.Storage,	
		RemoteConfig:   g.RemoteConfig,
	}

	return anyCredAPI.DeleteAnyCred(ctx, creds, pb.CredType_NOTIFIER)
}

func (g *NotifierSecretAPI) GetAnyCred(credder pb.OcyCredder) (pb.OcyCredder, error) {
	if credder.GetSubType() == 0 || credder.GetAcctName() == "" || credder.GetIdentifier() == "" {
		return nil, status.Error(codes.InvalidArgument, "subType, acctName, and identifier are required fields")
	}
	creddy, err := g.RemoteConfig.GetCred(g.Storage, credder.GetSubType(), credder.GetIdentifier(), credder.GetAcctName(), true)
	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); ok {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("Credential %s/%s of Type %s Not Found", credder.GetAcctName(), credder.GetIdentifier(), credder.GetSubType()))
		}
		if _, ok := err.(*pb.ValidationErr); ok {
			return nil, status.Error(codes.InvalidArgument, "Invalid arguments, error: "+err.Error())
		}
		return nil, status.Error(codes.Unavailable, "Credential interface not available, error: "+err.Error())
	}
	return creddy, nil
}

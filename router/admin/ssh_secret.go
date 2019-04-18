package admin

import (
	"context"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/level11consulting/ocelot/secret"
)

type SshSecret interface {
	UpdateSSHCreds(context.Context, *pb.SSHKeyWrapper) (*empty.Empty, error)
	GetSSHCred(context.Context, *pb.SSHKeyWrapper) (*pb.SSHKeyWrapper, error)
	SSHCredExists(context.Context, *pb.SSHKeyWrapper) (*pb.Exists, error)
	SetSSHCreds(context.Context, *pb.SSHKeyWrapper) (*empty.Empty, error)
	GetSSHCreds(context.Context, *empty.Empty) (*pb.SSHWrap, error)
	DeleteSSHCreds(context.Context, *pb.SSHKeyWrapper) (*empty.Empty, error)
}

func (g *OcelotServerAPI) UpdateSSHCreds(ctx context.Context, creds *pb.SSHKeyWrapper) (*empty.Empty, error) {
	return g.UpdateAnyCred(ctx, creds)
}

func (g *OcelotServerAPI) SSHCredExists(ctx context.Context, creds *pb.SSHKeyWrapper) (*pb.Exists, error) {
	return g.CheckAnyCredExists(ctx, creds)
}

func (g *OcelotServerAPI) SetSSHCreds(ctx context.Context, creds *pb.SSHKeyWrapper) (*empty.Empty, error) {
	if creds.SubType.Parent() != pb.CredType_SSH {
		return nil, status.Error(codes.InvalidArgument, "Subtype must be of ssh type: "+strings.Join(pb.CredType_SSH.SubtypesString(), " | "))
	}
	// no validation necessary, its a file upload

	err := secret.SetupRCCCredentials(g.DeprecatedHandler.RemoteConfig, g.DeprecatedHandler.Storage, creds)
	if err != nil {
		if _, ok := err.(*pb.ValidationErr); ok {
			return &empty.Empty{}, status.Error(codes.InvalidArgument, "SSH Creds Upload failed validation. Errors are: "+err.Error())
		}
		return &empty.Empty{}, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}

func (g *OcelotServerAPI) GetSSHCreds(context.Context, *empty.Empty) (*pb.SSHWrap, error) {
	credWrapper := &pb.SSHWrap{}
	credz, err := g.DeprecatedHandler.RemoteConfig.GetCredsByType(g.DeprecatedHandler.Storage, pb.CredType_SSH, true)
	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); ok {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return credWrapper, status.Errorf(codes.Internal, "unable to get ssh creds! error: %s", err.Error())
	}
	for _, v := range credz {
		credWrapper.Keys = append(credWrapper.Keys, v.(*pb.SSHKeyWrapper))
	}
	if len(credWrapper.Keys) == 0 {
		return credWrapper, status.Error(codes.NotFound, "no ssh keys found")
	}
	return credWrapper, nil
}

func (g *OcelotServerAPI) GetSSHCred(ctx context.Context, credentials *pb.SSHKeyWrapper) (*pb.SSHKeyWrapper, error) {
	creddy, err := g.GetAnyCred(credentials)
	if err != nil {
		return nil, err
	}
	ssh, ok := creddy.(*pb.SSHKeyWrapper)
	if !ok {
		return nil, status.Error(codes.Internal, "Unable to cast as SSH Creds")
	}
	return ssh, nil
}

func (g *OcelotServerAPI) DeleteSSHCreds(ctx context.Context, creds *pb.SSHKeyWrapper) (*empty.Empty, error) {
	return g.DeleteAnyCred(ctx, creds, pb.CredType_SSH)
}

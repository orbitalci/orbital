package admin

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/level11consulting/ocelot/models/pb"
	vaultkv "github.com/level11consulting/ocelot/server/config/vault"
	"github.com/level11consulting/ocelot/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type VcsSecret interface {
	GetVCSCreds(context.Context, *empty.Empty) (*pb.CredWrapper, error)
	GetVCSCred(context.Context, *pb.VCSCreds) (*pb.VCSCreds, error)
	SetVCSCreds(context.Context, *pb.VCSCreds) (*empty.Empty, error)
	UpdateVCSCreds(context.Context, *pb.VCSCreds) (*empty.Empty, error)
	VCSCredExists(context.Context, *pb.VCSCreds) (*pb.Exists, error)
    SetVCSPrivateKey(context.Context, *pb.SSHKeyWrapper) (*empty.Empty, error)
}

func (g *OcelotServerAPI) GetVCSCreds(ctx context.Context, msg *empty.Empty) (*pb.CredWrapper, error) {
	credWrapper := &pb.CredWrapper{}
	creds, err := g.RemoteConfig.GetCredsByType(g.Storage, pb.CredType_VCS, true)

	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); !ok {
			return credWrapper, status.Error(codes.NotFound, err.Error())
		}
		return credWrapper, status.Error(codes.Internal, "unable to get credentials, err: "+err.Error())
	}

	for _, v := range creds {
		vcsCred := v.(*pb.VCSCreds)
		sshKeyPath := vaultkv.BuildCredPath(vcsCred.SubType, vcsCred.AcctName, vcsCred.SubType.Parent(), v.GetIdentifier())
		err := g.RemoteConfig.CheckSSHKeyExists(sshKeyPath)
		if err != nil {
			vcsCred.SshFileLoc = "No SSH Key"
		} else {
			vcsCred.SshFileLoc = "SSH Key on file"
		}
		credWrapper.Vcs = append(credWrapper.Vcs, vcsCred)
	}
	if len(credWrapper.Vcs) == 0 {
		return nil, status.Error(codes.NotFound, "no vcs creds found")
	}
	return credWrapper, nil
}

func (g *OcelotServerAPI) SetVCSCreds(ctx context.Context, credentials *pb.VCSCreds) (*empty.Empty, error) {
	if credentials.SubType.Parent() != pb.CredType_VCS {
		return nil, status.Error(codes.InvalidArgument, "Subtype must be of vcs type: "+strings.Join(pb.CredType_VCS.SubtypesString(), " | "))
	}

	err := g.AdminValidator.ValidateConfig(credentials)
	if _, ok := err.(*pb.ValidationErr); ok {
		return &empty.Empty{}, status.Error(codes.InvalidArgument, "VCS Creds failed validation. Errors are: "+err.Error())
	}

	err = SetupCredentials(g, credentials)
	if err != nil {
		if err == unsupported {
			return nil, status.Error(codes.Unimplemented, "bitbucket is currently the only supported vcs type")
		}
		if er, ok := err.(*pb.ValidationErr); ok {
			return nil, status.Error(codes.InvalidArgument, er.Error())
		}
		// todo: make this a better error
		return &empty.Empty{}, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}

func (g *OcelotServerAPI) GetVCSCred(ctx context.Context, credentials *pb.VCSCreds) (*pb.VCSCreds, error) {
	creddy, err := g.GetAnyCred(credentials)
	if err != nil {
		return nil, err
	}
	vcs, ok := creddy.(*pb.VCSCreds)
	if !ok {
		return nil, status.Error(codes.Internal, "Unable to cast as VCS Creds")
	}
	return vcs, nil
}

func (g *OcelotServerAPI) GetAnyCred(credder pb.OcyCredder) (pb.OcyCredder, error) {
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

func (g *OcelotServerAPI) UpdateVCSCreds(ctx context.Context, credentials *pb.VCSCreds) (*empty.Empty, error) {
	credentials.Identifier = credentials.BuildIdentifier()
	return g.UpdateAnyCred(ctx, credentials)
}

func (g *OcelotServerAPI) VCSCredExists(ctx context.Context, credentials *pb.VCSCreds) (*pb.Exists, error) {
	credentials.Identifier = credentials.BuildIdentifier()
	return g.CheckAnyCredExists(ctx, credentials)
}
 
func (g *OcelotServerAPI) SetVCSPrivateKey(ctx context.Context, sshKeyWrapper *pb.SSHKeyWrapper) (*empty.Empty, error) {
	identifier, err := pb.CreateVCSIdentifier(sshKeyWrapper.SubType, sshKeyWrapper.AcctName)
	if err != nil {
		return &empty.Empty{}, status.Error(codes.InvalidArgument, err.Error())
	}
	sshKeyPath := vaultkv.BuildCredPath(sshKeyWrapper.SubType, sshKeyWrapper.AcctName, sshKeyWrapper.SubType.Parent(), identifier)
	err = g.RemoteConfig.AddSSHKey(sshKeyPath, sshKeyWrapper.PrivateKey)
	if err != nil {
		return &empty.Empty{}, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}

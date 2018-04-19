package admin

import (
	"context"
	"strings"

	cred "bitbucket.org/level11consulting/ocelot/common/credentials"
	"bitbucket.org/level11consulting/ocelot/models/pb"
	"bitbucket.org/level11consulting/ocelot/storage"
	"github.com/golang/protobuf/ptypes/empty"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)


func (g *guideOcelotServer) GetVCSCreds(ctx context.Context, msg *empty.Empty) (*pb.CredWrapper, error) {
	credWrapper := &pb.CredWrapper{}
	creds, err := g.RemoteConfig.GetCredsByType(g.Storage, pb.CredType_VCS, true)

	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); !ok {
			return credWrapper, err
		}
		return credWrapper, status.Error(codes.Internal, "unable to get credentials, err: " + err.Error())
	}

	for _, v := range creds {
		vcsCred := v.(*pb.VCSCreds)
		sshKeyPath := cred.BuildCredPath(vcsCred.SubType, vcsCred.AcctName, vcsCred.SubType.Parent(), v.GetIdentifier())
		err := g.RemoteConfig.CheckSSHKeyExists(sshKeyPath)
		if err != nil {
			vcsCred.SshFileLoc = "\033[0;33mNo SSH Key\033[0m"
		} else {
			vcsCred.SshFileLoc = "\033[0;34mSSH Key on file\033[0m"
		}
		credWrapper.Vcs = append(credWrapper.Vcs, vcsCred)
	}
	if len(credWrapper.Vcs) == 0 {
		return nil, status.Error(codes.NotFound, "no vcs creds found")
	}
	return credWrapper, nil
}


func (g *guideOcelotServer) SetVCSCreds(ctx context.Context, credentials *pb.VCSCreds) (*empty.Empty, error) {
	if credentials.SubType.Parent() != pb.CredType_VCS {
		return nil, status.Error(codes.InvalidArgument, "Subtype must be of vcs type: " + strings.Join(pb.CredType_VCS.SubtypesString(), " | "))
	}

	err := g.AdminValidator.ValidateConfig(credentials)
	if _, ok := err.(*pb.ValidationErr); ok {
		return &empty.Empty{}, status.Error(codes.InvalidArgument, "VCS Creds failed validation. Errors are: " + err.Error())
	}

	err = SetupCredentials(g, credentials)
	if err != nil {
		// todo: make this a better error
		return &empty.Empty{}, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}



func (g *guideOcelotServer) UpdateVCSCreds(ctx context.Context, credentials *pb.VCSCreds) (*empty.Empty, error) {
	credentials.Identifier = credentials.BuildIdentifier()
	return g.updateAnyCred(ctx, credentials)
}

func (g *guideOcelotServer) VCSCredExists(ctx context.Context, credentials *pb.VCSCreds) (*pb.Exists, error) {
	credentials.Identifier = credentials.BuildIdentifier()
	return g.checkAnyCredExists(ctx, credentials)
}

func (g *guideOcelotServer) GetRepoCreds(ctx context.Context, msg *empty.Empty) (*pb.RepoCredWrapper, error) {
	credWrapper := &pb.RepoCredWrapper{}
	creds, err := g.RemoteConfig.GetCredsByType(g.Storage, pb.CredType_REPO, true)

	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); !ok {
			return credWrapper, err
		}
	}

	for _, v := range creds {
		credWrapper.Repo = append(credWrapper.Repo, v.(*pb.RepoCreds))
	}
	if len(credWrapper.Repo) == 0 {
		return nil, status.Error(codes.NotFound, "no repo creds found")
	}
	return credWrapper, nil
}

func (g *guideOcelotServer) SetRepoCreds(ctx context.Context, creds *pb.RepoCreds) (*empty.Empty, error) {
	if creds.SubType.Parent() != pb.CredType_REPO {
		return nil, status.Error(codes.InvalidArgument, "Subtype must be of repo type: " + strings.Join(pb.CredType_REPO.SubtypesString(), " | "))
	}
	err := g.RepoValidator.ValidateConfig(creds)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed repo creds validation! error: %s", err.Error())
	}
	err = SetupRCCCredentials(g.RemoteConfig, g.Storage, creds)
	if _, ok := err.(*pb.ValidationErr); ok {
		return &empty.Empty{}, status.Error(codes.FailedPrecondition, "Repo Creds failed validation. Errors are: " + err.Error())
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

func (g *guideOcelotServer) SetK8SCreds(ctx context.Context, creds *pb.K8SCreds) (*empty.Empty, error) {
	if creds.SubType.Parent() != pb.CredType_K8S {
		return nil, status.Error(codes.InvalidArgument, "Subtype must be of k8s type: " + strings.Join(pb.CredType_K8S.SubtypesString(), " | "))
	}
	// no validation necessary, its a file upload

	err := SetupRCCCredentials(g.RemoteConfig,g.Storage, creds)
	if err != nil {
		// todo: make this better error
		return &empty.Empty{}, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}

func (g *guideOcelotServer) GetK8SCreds(ctx context.Context, empti *empty.Empty) (*pb.K8SCredsWrapper, error) {
	credWrapper := &pb.K8SCredsWrapper{}
	creds, err := g.RemoteConfig.GetCredsByType(g.Storage, pb.CredType_K8S, true)
	if err != nil {
		return credWrapper, status.Errorf(codes.Internal, "unable to get k8s creds! error: %s", err.Error())
	}
	for _, v := range creds {
		credWrapper.K8SCreds = append(credWrapper.K8SCreds, v.(*pb.K8SCreds))
	}
	if len(credWrapper.K8SCreds) == 0 {
		return credWrapper, status.Error(codes.NotFound, "no kubernetes integration creds found")
	}
	return credWrapper, nil
}

func (g *guideOcelotServer) UpdateK8SCreds(ctx context.Context, creds *pb.K8SCreds) (*empty.Empty, error) {
	return g.updateAnyCred(ctx, creds)
}

func (g *guideOcelotServer) K8SCredExists(ctx context.Context, creds *pb.K8SCreds) (*pb.Exists, error) {
	return g.checkAnyCredExists(ctx, creds)
}


func (g *guideOcelotServer) updateAnyCred(ctx context.Context, creds pb.OcyCredder) (*empty.Empty, error) {
	if err := g.RemoteConfig.UpdateCreds(g.Storage, creds); err != nil {
		if _, ok := err.(*pb.ValidationErr); ok {
			return &empty.Empty{}, status.Errorf(codes.FailedPrecondition, "%s cred failed validation. Errors are: %s", creds.GetSubType().Parent(), err.Error())
		}
		return &empty.Empty{}, status.Error(codes.Unavailable, err.Error())
	}
	return &empty.Empty{}, nil
}

func (g *guideOcelotServer) checkAnyCredExists(ctx context.Context, creds pb.OcyCredder) (*pb.Exists, error) {
	exists, err := g.Storage.CredExists(creds)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "Unable to reach cred table to check if cred %s/%s/%s exists. Error: %s", creds.GetAcctName(), creds.GetSubType().String(), creds.GetIdentifier(), err.Error())
	}
	return &pb.Exists{Exists:exists}, nil
}

func (g *guideOcelotServer) GetAllCreds(ctx context.Context, msg *empty.Empty) (*pb.AllCredsWrapper, error) {
	allCreds := &pb.AllCredsWrapper{}
	repoCreds, err := g.GetRepoCreds(ctx, msg)
	if err != nil {
		return allCreds, status.Errorf(codes.Internal, "unable to get repo creds! error: %s", err.Error())
	}
	allCreds.RepoCreds = repoCreds
	adminCreds, err := g.GetVCSCreds(ctx, msg)
	if err != nil {
		return allCreds, status.Errorf(codes.Internal, "unable to get vcs creds! error: %s", err.Error())
	}
	allCreds.VcsCreds = adminCreds
	return allCreds, nil
}


func (g *guideOcelotServer) SetVCSPrivateKey(ctx context.Context, sshKeyWrapper *pb.SSHKeyWrapper) (*empty.Empty, error) {
	identifier, err := pb.CreateVCSIdentifier(sshKeyWrapper.SubType, sshKeyWrapper.AcctName)
	if err != nil {
		return &empty.Empty{}, status.Error(codes.FailedPrecondition, err.Error())
	}
	sshKeyPath := cred.BuildCredPath(sshKeyWrapper.SubType, sshKeyWrapper.AcctName, sshKeyWrapper.SubType.Parent(), identifier)
	err = g.RemoteConfig.AddSSHKey(sshKeyPath, sshKeyWrapper.PrivateKey)
	if err != nil {
		return &empty.Empty{}, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}

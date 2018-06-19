package admin

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/shankj3/go-til/log"
	cred "github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/common/helpers/ioshelper"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (g *guideOcelotServer) GetVCSCreds(ctx context.Context, msg *empty.Empty) (*pb.CredWrapper, error) {
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
		sshKeyPath := cred.BuildCredPath(vcsCred.SubType, vcsCred.AcctName, vcsCred.SubType.Parent(), v.GetIdentifier())
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

func (g *guideOcelotServer) SetVCSCreds(ctx context.Context, credentials *pb.VCSCreds) (*empty.Empty, error) {
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

func (g *guideOcelotServer) GetVCSCred(ctx context.Context, credentials *pb.VCSCreds) (*pb.VCSCreds, error) {
	creddy, err := g.getAnyCred(credentials)
	if err != nil {
		return nil, err
	}
	vcs, ok := creddy.(*pb.VCSCreds)
	if !ok {
		return nil, status.Error(codes.Internal, "Unable to cast as VCS Creds")
	}
	return vcs, nil
}

func (g *guideOcelotServer) getAnyCred(credder pb.OcyCredder) (pb.OcyCredder, error){
	if credder.GetSubType() == 0 || credder.GetAcctName() ==  "" || credder.GetIdentifier() == "" {
		return nil, status.Error(codes.InvalidArgument, "subType, acctName, and identifier are required fields")
	}
	creddy, err := g.RemoteConfig.GetCred(g.Storage, credder.GetSubType(), credder.GetIdentifier(), credder.GetAcctName(), true)
	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); ok {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("Credential %s/%s of Type %s Not Found", credder.GetAcctName(), credder.GetIdentifier(), credder.GetSubType()))
		}
		if _, ok := err.(*pb.ValidationErr); ok {
			return nil, status.Error(codes.InvalidArgument, "Invalid arguments, error: " + err.Error())
		}
		return nil, status.Error(codes.Unavailable, "Credential interface not available, error: " + err.Error())
	}
	return creddy, nil
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
		return &empty.Empty{}, status.Error(codes.FailedPrecondition, "Repo Creds failed validation. Errors are: "+err.Error())
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
		return nil, status.Error(codes.InvalidArgument, "Subtype must be of k8s type: "+strings.Join(pb.CredType_K8S.SubtypesString(), " | "))
	}
	// no validation necessary, its a file upload

	err := SetupRCCCredentials(g.RemoteConfig, g.Storage, creds)
	if err != nil {
		if _, ok := err.(*pb.ValidationErr); ok {
			return &empty.Empty{}, status.Error(codes.FailedPrecondition, "K8s Creds failed validation. Errors are: "+err.Error())
		}
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


func (g *guideOcelotServer) GetK8SCred(ctx context.Context, credentials *pb.K8SCreds) (*pb.K8SCreds, error) {
	creddy, err := g.getAnyCred(credentials)
	if err != nil {
		return nil, err
	}
	repo, ok := creddy.(*pb.K8SCreds)
	if !ok {
		return nil, status.Error(codes.Internal, "Unable to cast as Kubernetes Creds")
	}
	return repo, nil
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
			return &empty.Empty{}, status.Errorf(codes.InvalidArgument, "%s cred failed validation. Errors are: %s", creds.GetSubType().Parent(), err.Error())
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
	return &pb.Exists{Exists: exists}, nil
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

func (g *guideOcelotServer) UpdateSSHCreds(ctx context.Context, creds *pb.SSHKeyWrapper) (*empty.Empty, error) {
	return g.updateAnyCred(ctx, creds)
}

func (g *guideOcelotServer) SSHCredExists(ctx context.Context, creds *pb.SSHKeyWrapper) (*pb.Exists, error) {
	return g.checkAnyCredExists(ctx, creds)
}

func (g *guideOcelotServer) SetSSHCreds(ctx context.Context, creds *pb.SSHKeyWrapper) (*empty.Empty, error) {
	if creds.SubType.Parent() != pb.CredType_SSH {
		return nil, status.Error(codes.InvalidArgument, "Subtype must be of ssh type: "+strings.Join(pb.CredType_SSH.SubtypesString(), " | "))
	}
	// no validation necessary, its a file upload

	err := SetupRCCCredentials(g.RemoteConfig, g.Storage, creds)
	if err != nil {
		if _, ok := err.(*pb.ValidationErr); ok {
			return &empty.Empty{}, status.Error(codes.InvalidArgument, "SSH Creds Upload failed validation. Errors are: "+err.Error())
		}
		return &empty.Empty{}, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}

func (g *guideOcelotServer) GetSSHCreds(context.Context, *empty.Empty) (*pb.SSHWrap, error) {
	credWrapper := &pb.SSHWrap{}
	credz, err := g.RemoteConfig.GetCredsByType(g.Storage, pb.CredType_SSH, true)
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

func (g *guideOcelotServer) GetSSHCred(ctx context.Context, credentials *pb.SSHKeyWrapper) (*pb.SSHKeyWrapper, error) {
	creddy, err := g.getAnyCred(credentials)
	if err != nil {
		return nil, err
	}
	ssh, ok := creddy.(*pb.SSHKeyWrapper)
	if !ok {
		return nil, status.Error(codes.Internal, "Unable to cast as SSH Creds")
	}
	return ssh, nil
}

func appleNastiness(zipFile []byte, devProfilePassword string) (parsed []byte, err error) {
	appleKeychain, err := ioshelper.UnpackAppleDevAccount(zipFile, devProfilePassword)
	if err != nil {
		log.IncludeErrField(err).Error("couldn't deal with this zip file...")
		return nil, status.Error(codes.InvalidArgument, "could not unpack developeraccount zip to keychain, error is :" + err.Error())
	}
	return appleKeychain, nil
}


func (g *guideOcelotServer) SetAppleCreds(ctx context.Context, creds *pb.AppleCreds) (*empty.Empty, error) {
	vempty := &empty.Empty{}
	if creds.GetSubType().Parent() != pb.CredType_APPLE {
		return nil, status.Error(codes.InvalidArgument, "Subtype must be of apple type: " + strings.Join(pb.CredType_APPLE.SubtypesString(), " | "))
	}
	var err error
	creds.AppleSecrets, err = appleNastiness(creds.AppleSecrets, creds.AppleSecretsPassword)
	if err != nil {
		return nil, err
	}

	if err := SetupRCCCredentials(g.RemoteConfig, g.Storage, creds); err != nil {
		if _, ok := err.(*pb.ValidationErr); ok {
			return vempty, status.Error(codes.InvalidArgument, "Apple creds upload failed validation, errors are: " + err.Error())
		}
		return vempty, status.Error(codes.Internal, "Apple creds could not be uploaded, error is: " + err.Error())
	}
	log.Log().Info("unpacked & stored apple dev profile")
	return vempty, nil
}

func (g *guideOcelotServer) GetAppleCreds(ctx context.Context, empty2 *empty.Empty) (*pb.AppleCredsWrapper, error) {
	wrapper := &pb.AppleCredsWrapper{}
	credz, err := g.RemoteConfig.GetCredsByType(g.Storage, pb.CredType_APPLE, true)
	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); ok {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return wrapper, status.Errorf(codes.Internal, "unable to get apple creds! error: %s", err.Error())
	}
	for _, v := range credz {
		wrapper.AppleCreds = append(wrapper.AppleCreds, v.(*pb.AppleCreds))
	}
	if len(wrapper.AppleCreds) == 0 {
		return nil, status.Error(codes.NotFound, "no apple creds found")
	}
	return wrapper, nil
}


func (g *guideOcelotServer) GetAppleCred(ctx context.Context, creds *pb.AppleCreds) (*pb.AppleCreds, error){
	creddy, err := g.getAnyCred(creds)
	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); ok {
			return nil, status.Error(codes.NotFound, "apple cred not found " + err.Error())
		}
		return nil, status.Error(codes.Internal, "unexpected error occured, apple creds could not be retrieved. error is: " + err.Error())
	}
	apple, ok := creddy.(*pb.AppleCreds)
	if !ok {
		return nil, status.Error(codes.Internal, "unable to cast as apple creds")
	}
	return apple, nil
}

func (g *guideOcelotServer) UpdateAppleCreds(ctx context.Context, creds *pb.AppleCreds) (*empty.Empty, error) {
	var err error
	creds.AppleSecrets, err = appleNastiness(creds.AppleSecrets, creds.AppleSecretsPassword)
	if err != nil {
		return nil, err
	}
	return g.updateAnyCred(ctx, creds)
}

func (g *guideOcelotServer) AppleCredExists(ctx context.Context, creds *pb.AppleCreds) (*pb.Exists, error) {
	return g.checkAnyCredExists(ctx, creds)
}
/*SetNotifyCreds(context.Context, *NotifyCreds) (*google_protobuf.Empty, error)
	GetNotifyCred(context.Context, *NotifyCreds) (*NotifyCreds, error)
	UpdateNotifyCreds(context.Context, *NotifyCreds) (*google_protobuf.Empty, error)
	NotifyCredExists(context.Context, *NotifyCreds) (*Exists, error)
*/

func (g *guideOcelotServer) SetNotifyCreds(ctx context.Context, creds *pb.NotifyCreds) (*empty.Empty, error) {
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

func (g *guideOcelotServer) NotifyCredExists(ctx context.Context, creds *pb.NotifyCreds) (*pb.Exists, error) {
	return g.checkAnyCredExists(ctx, creds)
}

func (g *guideOcelotServer) UpdateNotifyCreds(ctx context.Context, creds *pb.NotifyCreds) (*empty.Empty, error) {
	return g.updateAnyCred(ctx, creds)
}

func (g *guideOcelotServer) GetNotifyCred(ctx context.Context, creds *pb.NotifyCreds) (*pb.NotifyCreds, error) {
	creddy, err := g.getAnyCred(creds)
	if err != nil {
		return nil, err
	}
	notifier, ok := creddy.(*pb.NotifyCreds)
	if !ok {
		return nil, status.Error(codes.Internal, "Unable to cast as Notifier Creds")
	}
	return notifier, nil
}

func (g *guideOcelotServer) GetNotifyCreds(ctx context.Context, empty2 *empty.Empty) (*pb.NotifyWrap, error) {
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

func (g *guideOcelotServer) GetGenericCreds(ctx context.Context, empty *empty.Empty) (*pb.GenericWrap, error) {
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

func (g *guideOcelotServer) SetGenericCreds(ctx context.Context, wrap *pb.GenericWrap) (*empty.Empty, error) {
	for _, creds := range wrap.Creds {
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
	}
	return &empty.Empty{}, nil
}

func (g *guideOcelotServer) GenericCredExists(ctx context.Context, creds *pb.GenericCreds) (*pb.Exists, error) {
	return g.checkAnyCredExists(ctx, creds)
}

func (g *guideOcelotServer) UpdateGenericCreds(ctx context.Context, creds *pb.GenericCreds) (*empty.Empty, error) {
	return g.updateAnyCred(ctx, creds)
}
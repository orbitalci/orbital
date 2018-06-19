package admin

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/shankj3/go-til/test"
	"github.com/shankj3/ocelot/common/credentials"
	"github.com/shankj3/ocelot/common/helpers/ioshelper"
	"github.com/shankj3/ocelot/models/pb"
	"github.com/shankj3/ocelot/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGuideOcelotServer_GetVCSCreds(t *testing.T) {
	rc := &vcsRemoteConf{}
	ctx := context.Background()
	gos := &guideOcelotServer{RemoteConfig:rc}
	wrap, err := gos.GetVCSCreds(ctx, nil)
	if err != nil {
		t.Error(err)
	}
	if len(wrap.Vcs) != 3 {
		t.Error("something went awry")
	}
	rc.sshExists = true
	wrap, err = gos.GetVCSCreds(ctx, nil)
	if err != nil {
		t.Error(err)
	}
	if wrap.Vcs[0].SshFileLoc != "SSH Key on file" {
		t.Error("ssh key exists, should reflect in the SshFileLoc field")
	}
	rc.empty = true
	_, err = gos.GetVCSCreds(ctx, nil)
	if err == nil {
		t.Error("no creds returend should return an error")
	}
	statusErr, ok := status.FromError(err)
	if !ok {
		t.Error("admin should return only grpc errors")
	}
	if statusErr.Code() != codes.NotFound {
		t.Error("should return not found if no creds returned. ")
	}
	rc.empty = false
	rc.notFound = true
	_, err = gos.GetVCSCreds(ctx, nil)
	if err == nil {
		t.Error("remote conf returned a storage.NotFound error, should bubble up")
	}
	rc.notFound = false
	rc.returnErr = true
	_, err = gos.GetVCSCreds(ctx, nil)
	if err == nil {
		t.Error("GetCredsByType returned an unhandled error, shoul bubble up")
	}

}


func TestGuideOcelotServer_SetVCSCreds(t *testing.T) {
	rc := &vcsRemoteConf{}
	ctx := context.Background()
	gos := &guideOcelotServer{RemoteConfig:rc, AdminValidator: & credentials.AdminValidator{}}
	cred := &pb.VCSCreds{
		SubType: pb.SubCredType_BITBUCKET,
		ClientId: "hi",
		ClientSecret: "secret",
		AcctName: "shankj3",
		Identifier: "identifier",
		TokenURL: "http://adsfadsfasdfasdfadsfasdfasdfadsfasdfasdfasdfasdfasdf.com",
	}
	// even tho the url is bunk, it last calls the remoteconfig which we can control
	_, err := gos.SetVCSCreds(ctx, cred)
	if err != nil {
		t.Error("should work fine")
	}
	cred.SubType = pb.SubCredType_DEVPROFILE
	_, err = gos.SetVCSCreds(ctx, cred)
	if err == nil {
		t.Error("wrong subtype, should fial")
	}
	cred.SubType = pb.SubCredType_GITHUB
	_, err = gos.SetVCSCreds(ctx, cred)
	if err == nil {
		t.Error("unsuported vcs, should fial")
	}
	cred.SubType = pb.SubCredType_BITBUCKET
	rc.validationErr = true
	_, err = gos.SetVCSCreds(ctx, cred)
	if err == nil {
		t.Error("failed validation, should fail")
	}
	rc.validationErr = false
	rc.returnErr = true
	_, err = gos.SetVCSCreds(ctx, cred)
	if err == nil {
		t.Error("unhandled error, should fail")
	}

}

func TestGuideOcelotServer_GetVCSCred(t *testing.T) {
	rc := &vcsRemoteConf{}
	ctx := context.Background()
	gos := &guideOcelotServer{RemoteConfig:rc}
	_, err := gos.GetVCSCred(ctx, &pb.VCSCreds{ClientId:"xxxx", Identifier:"id", AcctName:"account", SubType: pb.SubCredType_BITBUCKET})
	if err != nil {
		t.Error(err)
	}
	rc.returnWrong = true
	_, err = gos.GetVCSCred(ctx, &pb.VCSCreds{ClientId:"xxxx", Identifier:"id", AcctName:"account", SubType: pb.SubCredType_BITBUCKET})
	if err == nil {
		t.Error("returned wrong struct, should fail")
	}
}

func TestGuideOcelotServer_getAnyCred(t *testing.T) {
	rc := &vcsRemoteConf{}
	gos := &guideOcelotServer{RemoteConfig:rc}
	rc.notFound = true
	_, err := gos.getAnyCred(&pb.VCSCreds{AcctName:"hi", Identifier:"123", SubType:pb.SubCredType_BITBUCKET})
	if err == nil {
		t.Error("remote config returned a not found, this shoudl fail")
	}
	statErr, ok := status.FromError(err)
	if !ok {
		t.Error("admin router should only return grpc status errors")
		return
	}
	if statErr.Code() != codes.NotFound {
		t.Error("should return grpc code of not found if remoteconfig returns a NotFound errror")
	}
	if statErr.Message() !=  "Credential hi/123 of Type BITBUCKET Not Found" {
		t.Error("should return not found message, got: " + statErr.Message())
	}
	rc.notFound = false
	rc.validationErr = true
	_, err = gos.getAnyCred(&pb.VCSCreds{AcctName:"hi", Identifier:"123", SubType:pb.SubCredType_BITBUCKET})
	if err == nil {
		t.Error("remote config returned a validation error, this shoudl fail")
	}
	statErr, ok = status.FromError(err)
	if !ok {
		t.Error("admin router should only return grpc status errors")
		return
	}
	if statErr.Code() != codes.InvalidArgument {
		t.Error("validation errors shoudl return InvalidArgument code")
	}
	rc.validationErr = false
	rc.returnErr = true
	_, err = gos.getAnyCred(&pb.VCSCreds{AcctName:"hi", Identifier:"123", SubType:pb.SubCredType_BITBUCKET})
	if err == nil {
		t.Error("remote config returned a validation error, this shoudl fail")
	}
	statErr, ok = status.FromError(err)
	if !ok {
		t.Error("admin router should only return grpc status errors")
		return
	}
	if statErr.Code() != codes.Unavailable {
		t.Error("remote config returned an unhandled error, server should return unavailable")
	}
	rc.returnErr = false
	_, err = gos.getAnyCred(&pb.VCSCreds{AcctName:"hi", Identifier:"123", SubType:pb.SubCredType_BITBUCKET})
	if err != nil {
		t.Error("remote config is playing nice, this should just work. instead error is: " + err.Error())
	}

}

func TestGuideOcelotServer_updateAnyCred(t *testing.T) {
	rc := &vcsRemoteConf{}
	gos := &guideOcelotServer{RemoteConfig:rc}
	var err error
	ctx := context.Background()
	cred := &pb.VCSCreds{}
	_, err = gos.updateAnyCred(ctx, cred)
	if err != nil {
		t.Error("remote conf is flagged to play nice, this should execute successfully")
	}
	rc.validationErr = true
	_, err = gos.updateAnyCred(ctx, cred)
	if err == nil {
		t.Error("remoteconf returned a validationErr, should bubble up")
	}
	statusErr, ok := status.FromError(err)
	if !ok {
		t.Error("admin errors should be grpc errors")
	}
	if statusErr.Code() != codes.InvalidArgument {
		t.Error("credential failed validation, return code should be invalid argument")
	}
	rc.validationErr = false
	rc.returnErr = true
	_, err = gos.updateAnyCred(ctx, cred)
	if err == nil {
		t.Error("remoteconf returned a validationErr, should bubble up")
	}
	statusErr, ok = status.FromError(err)
	if !ok {
		t.Error("admin errors should be grpc errors")
	}
	if statusErr.Code() != codes.Unavailable {
		t.Error("remote conf returned an unidentifiable error, should return a code of unavailable")
	}
}

func TestGuideOcelotServer_UpdateVCSCreds(t *testing.T) {
	rc := &vcsRemoteConf{}
	gos := &guideOcelotServer{RemoteConfig:rc}
	cred := &pb.VCSCreds{AcctName:"shankj3", SubType:pb.SubCredType_BITBUCKET}
	if _, err := gos.UpdateVCSCreds(context.Background(), cred); err != nil {
		t.Error("remoteconf is playing nice, this hsould not fail")
	}
	if rc.updated[0].(*pb.VCSCreds).Identifier != "BITBUCKET_shankj3" {
		t.Error("idnetiifer should be BITBUCKET_shankj3, got " + rc.updated[0].(*pb.VCSCreds).Identifier)
	}
}

func TestGuideOcelotServer_checkAnyCredExists(t *testing.T) {
	rc := &vcsRemoteConf{}
	stor := &store{}
	gos := &guideOcelotServer{RemoteConfig:rc, Storage:stor}
	exists, err := gos.checkAnyCredExists(context.Background(), &pb.VCSCreds{})
	if err != nil {
		t.Error("should not return an error, as storage is reachable")
	}
	if exists.Exists {
		t.Error("should not exist as flag has not been set on *store")
	}
	stor.exists = true
	exists, err = gos.checkAnyCredExists(context.Background(), &pb.VCSCreds{})
	if err != nil {
		t.Error("should not return an error, as storage is reachable")
	}
	if !exists.Exists {
		t.Error("should exist as flag has been set on *store")
	}
	stor.exists = false
	stor.returnErr = true
	_, err = gos.checkAnyCredExists(context.Background(), &pb.VCSCreds{})
	if err == nil {
		t.Error("storage returned an error, this should return an error")
	}
}


func TestGuideOcelotServer_VCSCredExists(t *testing.T) {
	rc := &vcsRemoteConf{}
	stor := &store{exists:true}
	gos := &guideOcelotServer{RemoteConfig:rc, Storage:stor}
	exists, err := gos.VCSCredExists(context.Background(), &pb.VCSCreds{})
	if err != nil {
		t.Error("should pass")
	}
	if !exists.Exists {
		t.Error("should return  existing")
	}
}

func TestGuideOcelotServer_GetRepoCreds(t *testing.T) {
	rc := &vcsRemoteConf{}
	gos := &guideOcelotServer{RemoteConfig:rc}
	rc.notFound = true
	_, err := gos.GetRepoCreds(context.Background(), nil)
	if err == nil {
		t.Error("remote config returned a not found, this shoudl error")
	}
	rc.notFound = false
	creds, err := gos.GetRepoCreds(context.Background(), nil)
	if err != nil {
		t.Error("remote config behaved, this should work, err: " + err.Error())
	}
	if len(creds.Repo) != 3 {
		t.Error("wtf?")
	}
	rc.returnErr = true
	_, err = gos.GetRepoCreds(context.Background(), nil)
	if err == nil {
		t.Error("remote config returned an error, this shoudl error")
	}
	statusErr, ok := status.FromError(err)
	if !ok {
		t.Error("admin should return grpc errors. err is: " + err.Error())
	}
	if statusErr.Code() != codes.ResourceExhausted {
		t.Error("unidentfied error should return resource exhausted code")
	}
}

func TestGuideOcelotServer_GetRepoCred(t *testing.T) {
	rc := &vcsRemoteConf{}
	gos := &guideOcelotServer{RemoteConfig:rc}
	_, err := gos.GetRepoCred(context.Background(), &pb.RepoCreds{Identifier:"id", AcctName:"shankj3", SubType:pb.SubCredType_NEXUS})
	if err != nil {
		t.Error("should succeed, err: " + err.Error())
	}
	rc.returnErr = true
	_, err = gos.GetRepoCred(context.Background(), &pb.RepoCreds{Identifier:"id", AcctName:"shankj3", SubType:pb.SubCredType_NEXUS})
	if err == nil {
		t.Error("remote conf returned an error, this should bubble up")
	}
	rc.returnErr = false
	rc.returnWrong = true
	_, err = gos.GetRepoCred(context.Background(), &pb.RepoCreds{Identifier:"id", AcctName:"shankj3", SubType:pb.SubCredType_NEXUS})
	if err == nil {
		t.Error("returned wrong struct, should fail")
	}
}

func TestGuideOcelotServer_SetRepoCreds(t *testing.T) {
	rc := &vcsRemoteConf{}
	gos := &guideOcelotServer{RemoteConfig:rc, RepoValidator: credentials.GetRepoValidator()}
	creds := &pb.RepoCreds{
		Username: "user",
		Password: "password",
		RepoUrl: "http://repo.url",
		AcctName: "shankj3",
		Identifier: "id!me",
		SubType: pb.SubCredType_DOCKER,
	}
	ctx := context.Background()
	_, err := gos.SetRepoCreds(ctx, creds)
	if err != nil {
		t.Error("cred is valid and remote conf is set to not fail, this should work. error is: " + err.Error())
	}
	creds.RepoUrl = ""
	_, err = gos.SetRepoCreds(ctx, creds)
	if err == nil {
		t.Error("repourl is empty, this should fail validation")
	}
	if statusErr, ok := status.FromError(err); !ok {
		t.Error("admin shoudl return grpc errors")
	} else if statusErr.Code() != codes.InvalidArgument {
		t.Error("cred failed validation, should return invalid argument ")
	}
	creds.RepoUrl = "http://repo.url"
	creds.SubType = pb.SubCredType_GITHUB
	_, err = gos.SetRepoCreds(ctx, creds)
	if err == nil {
		t.Error("sub type does not match repo cred")
	}
	if statusErr, ok := status.FromError(err); !ok {
		t.Error("admin shoudl return grpc errors")
	} else if statusErr.Code() != codes.InvalidArgument {
		t.Error("cred is not repo cred, should return invalid argument ")
	}
	creds.SubType = pb.SubCredType_DOCKER

	rc.validationErr = true
	_, err = gos.SetRepoCreds(ctx, creds)
	if err == nil {
		t.Error("remote config returned a vlaidationerr, this should bubble up")
	}
	if statusErr, ok := status.FromError(err); !ok {
		t.Error("admin shoudl return grpc errors")
	} else if statusErr.Code() != codes.InvalidArgument {
		t.Error("cred failed remote conf validation, should return invalid argument ")
	}
	rc.validationErr = false
	rc.returnErr = true
	_, err = gos.SetRepoCreds(ctx, creds)
	if err == nil {
		t.Error("remote config returned an unknown error, this should bubble up")
	}
}

func TestGuideOcelotServer_UpdateRepoCreds(t *testing.T) {
	// this is tested heavily in updateAnyCred
	rc := &vcsRemoteConf{}
	gos := &guideOcelotServer{RemoteConfig:rc, RepoValidator: credentials.GetRepoValidator()}
	_, err := gos.UpdateRepoCreds(context.Background(), &pb.RepoCreds{})
	if err != nil {
		t.Error(err)
	}
}

func TestGuideOcelotServer_RepoCredExists(t *testing.T) {
	rc := &vcsRemoteConf{}
	stor := &store{exists:true}
	gos := &guideOcelotServer{RemoteConfig:rc, Storage:stor}
	exists, err := gos.RepoCredExists(context.Background(), &pb.RepoCreds{})
	if err != nil {
		t.Error(err)
	}
	if !exists.Exists {
		t.Error("remote conf returned that cred exists, should return that it exists")
	}
}

func TestGuideOcelotServer_SetK8SCreds(t *testing.T) {
	rc := &vcsRemoteConf{}
	gos := &guideOcelotServer{RemoteConfig:rc, RepoValidator: credentials.GetRepoValidator()}
	creds := &pb.K8SCreds{
		K8SContents:"hi",
		AcctName:"hi",
		Identifier:"hi",
		SubType: pb.SubCredType_KUBECONF,
	}
	ctx := context.Background()
	_, err := gos.SetK8SCreds(ctx, creds)
	if err != nil {
		t.Error(err)
	}
	rc.validationErr = true
	_, err = gos.SetK8SCreds(ctx, creds)
	if err == nil {
		t.Error("failed validation, should fail")
	}
	rc.validationErr = false
	creds.SubType = pb.SubCredType_DOCKER
	_, err = gos.SetK8SCreds(ctx, creds)
	if err  == nil {
		t.Error("not the right subtype, shoud fail")
	}
	creds.SubType = pb.SubCredType_KUBECONF
	rc.returnErr = true
	_, err = gos.SetK8SCreds(ctx, creds)
	if err == nil {
		t.Error("remote conf sent generic error, this should fail")
	}
}

func TestGuideOcelotServer_GetK8SCreds(t *testing.T) {
	rc := &vcsRemoteConf{}
	gos := &guideOcelotServer{RemoteConfig:rc, RepoValidator: credentials.GetRepoValidator()}
	ctx := context.Background()
	creds, err := gos.GetK8SCreds(ctx, nil)
	if err != nil {
		t.Error(err)
	}
	if len(creds.K8SCreds) != 3 {
		t.Error("not all creds retrieved? wtf?")
	}
	rc.empty = true
	_, err = gos.GetK8SCreds(ctx, nil)
	if err == nil {
		t.Error("no creds returned, should return not found")
	}
	staterr, ok := status.FromError(err)
	if !ok {
		t.Error("should be grpc err")
	}
	if staterr.Code() != codes.NotFound {
		t.Error("should be code not found, got: " + staterr.Code().String())
	}
	rc.empty = false
	rc.returnErr = true
	_, err = gos.GetK8SCreds(ctx, nil)
	if err == nil {
		t.Error("remote conf returned unidentified error, this should fail")
	}
	staterr, ok = status.FromError(err)
	if !ok {
		t.Error("should be grpc err")
	}
	if staterr.Code() != codes.Internal {
		t.Error("should return internal grpc status gode, got ", staterr.Code().String())
	}
}

func TestGuideOcelotServer_getK8sCred(t *testing.T) {
	cred := &pb.K8SCreds{AcctName:"shankj3", Identifier: "hi", SubType: pb.SubCredType_KUBECONF}
	rc := &vcsRemoteConf{}
	gos := &guideOcelotServer{RemoteConfig:rc}
	ctx := context.Background()
	cred, err := gos.GetK8SCred(ctx, cred)
	if err != nil {
		t.Error(err)
	}
	rc.returnErr = true
	_, err = gos.GetK8SCred(ctx, cred)
	if err == nil {
		t.Error("should bubble up remote conf get cred error")
	}
	rc.returnWrong = true
	rc.returnErr = false
	_, err = gos.GetK8SCred(ctx, cred)
	if err == nil {
		t.Error("returned wrong cred type, should error")
	}
}


func TestGuideOcelotServer_updateK8sCreds(t *testing.T) {
	// this is tested heavily in updateAnyCred
	rc := &vcsRemoteConf{}
	gos := &guideOcelotServer{RemoteConfig:rc,}
	_, err := gos.UpdateK8SCreds(context.Background(), &pb.K8SCreds{})
	if err != nil {
		t.Error(err)
	}
}



func TestGuideOcelotServer_K8SCredExists(t *testing.T) {
	// this is tested heavily in checkAnyCredExists
	rc := &vcsRemoteConf{}
	stor := &store{exists:true}
	gos := &guideOcelotServer{RemoteConfig:rc, Storage:stor}
	exists, err := gos.K8SCredExists(context.Background(), &pb.K8SCreds{})
	if err != nil {
		t.Error(err)
	}
	if !exists.Exists {
		t.Error("remote conf returned that cred exists, should return that it exists")
	}
}

func TestGuideOcelotServer_GetAllCreds(t *testing.T) {
	rc := &vcsRemoteConf{}
	stor := &store{exists:true}
	gos := &guideOcelotServer{RemoteConfig:rc, Storage:stor}
	all, err := gos.GetAllCreds(context.Background(), nil)
	if err != nil {
		t.Error(err)
	}
	if len(all.RepoCreds.Repo) != 3 {
		t.Error("wrong")
	}
	if len(all.VcsCreds.Vcs) != 3 {
		t.Error("wrong")
	}
}

func TestGuideOcelotServer_SetVCSPrivateKey(t *testing.T) {
	rc := &vcsRemoteConf{}
	gos := &guideOcelotServer{RemoteConfig:rc}
	cred := &pb.SSHKeyWrapper{
		AcctName: "shankj3",
		Identifier: "kubeconf",
		SubType: pb.SubCredType_BITBUCKET,
		PrivateKey: []byte("so priv much secret"),
	}
	ctx := context.Background()
	_, err := gos.SetVCSPrivateKey(ctx, cred)
	if err != nil {
		t.Error(err)
	}
	rc.returnErr = true
	_, err = gos.SetVCSPrivateKey(ctx, cred)
	if err == nil {
		t.Error("remote conf returned err, this should be bubbled up")
	}
	rc.returnErr = false
	cred.SubType = pb.SubCredType_SSHKEY
	_, err = gos.SetVCSPrivateKey(ctx, cred)
	if err == nil {
		t.Error("wrong sub type, should return error")
	}
}


func TestGuideOcelotServer_UpdateSSHCreds(t *testing.T) {
	// this is tested heavily in updateAnyCred
	rc := &vcsRemoteConf{}
	gos := &guideOcelotServer{RemoteConfig:rc, RepoValidator: credentials.GetRepoValidator()}
	_, err := gos.UpdateSSHCreds(context.Background(), &pb.SSHKeyWrapper{})
	if err != nil {
		t.Error(err)
	}
}

func TestGuideOcelotServer_SSHCredExists(t *testing.T) {
	// tested heavily in checkAnyCredExists
	rc := &vcsRemoteConf{}
	stor := &store{exists:true}
	gos := &guideOcelotServer{RemoteConfig:rc, Storage:stor}
	exists, err := gos.SSHCredExists(context.Background(), &pb.SSHKeyWrapper{})
	if err != nil {
		t.Error("should pass")
	}
	if !exists.Exists {
		t.Error("should return not existing")
	}
}

func TestGuideOcelotServer_SetSSHCreds(t *testing.T) {
	rc := &vcsRemoteConf{}
	gos := &guideOcelotServer{RemoteConfig:rc, RepoValidator: credentials.GetRepoValidator()}
	creds := &pb.SSHKeyWrapper{
		AcctName:"hi",
		Identifier:"hi",
		SubType: pb.SubCredType_SSHKEY,
		PrivateKey: []byte("much priv"),
	}
	ctx := context.Background()
	_, err := gos.SetSSHCreds(ctx, creds)
	if err != nil {
		t.Error(err)
	}
	rc.validationErr = true
	_, err = gos.SetSSHCreds(ctx, creds)
	if err == nil {
		t.Error("failed validation, should fail")
	}
	rc.validationErr = false
	creds.SubType = pb.SubCredType_DOCKER
	_, err = gos.SetSSHCreds(ctx, creds)
	if err  == nil {
		t.Error("not the right subtype, shoud fail")
	}
	creds.SubType = pb.SubCredType_SSHKEY
	rc.returnErr = true
	_, err = gos.SetSSHCreds(ctx, creds)
	if err == nil {
		t.Error("remote conf sent generic error, this should fail")
	}
}

func TestGuideOcelotServer_GetSSHCreds(t *testing.T) {
	rc := &vcsRemoteConf{}
	gos := &guideOcelotServer{RemoteConfig:rc, RepoValidator: credentials.GetRepoValidator()}
	ctx := context.Background()
	creds, err := gos.GetSSHCreds(ctx, nil)
	if err != nil {
		t.Error(err)
	}
	if len(creds.Keys) != 3 {
		t.Error("not all creds retrieved? wtf?")
	}
	rc.empty = true
	_, err = gos.GetSSHCreds(ctx, nil)
	if err == nil {
		t.Error("no creds returned, should return not found")
	}
	staterr, ok := status.FromError(err)
	if !ok {
		t.Error("should be grpc err")
	}
	if staterr.Code() != codes.NotFound {
		t.Error("should be code not found, got: " + staterr.Code().String())
	}
	rc.empty = false
	rc.returnErr = true
	_, err = gos.GetSSHCreds(ctx, nil)
	if err == nil {
		t.Error("remote conf returned unidentified error, this should fail")
	}
	staterr, ok = status.FromError(err)
	if !ok {
		t.Error("should be grpc err")
	}
	if staterr.Code() != codes.Internal {
		t.Error("should return internal grpc status gode, got ", staterr.Code().String())
	}
}

func TestGuideOcelotServer_GetSSHCred(t *testing.T) {
	cred := &pb.SSHKeyWrapper{AcctName:"shankj3", Identifier: "hi", SubType: pb.SubCredType_SSHKEY}
	rc := &vcsRemoteConf{}
	gos := &guideOcelotServer{RemoteConfig:rc}
	ctx := context.Background()
	cred, err := gos.GetSSHCred(ctx, cred)
	if err != nil {
		t.Error(err)
	}
	rc.returnErr = true
	_, err = gos.GetSSHCred(ctx, cred)
	if err == nil {
		t.Error("should bubble up remote conf get cred error")
	}
	rc.returnErr = false
	rc.returnWrong = true
	_, err = gos.GetSSHCred(ctx, cred)
	if err == nil {
		t.Error("returned wrong cred type, should fail")
	}
}

func TestGuideOcelotServer_appleNastiness(t *testing.T) {
	zipdata, pw := ioshelper.GetZipAndPw(t)
	_, err := appleNastiness(zipdata, pw)
	if err != nil {
		t.Error(err)
	}
	bunk := []byte("this is not a zip file, clearly")
	_, err = appleNastiness(bunk, pw)
	if err == nil {
		t.Error("not a real zip file, this should fail")
	}
}

func TestGuideOcelotServer_SetAppleCreds(t *testing.T) {
	zipdata, pw := ioshelper.GetZipAndPw(t)
	creds := &pb.AppleCreds{
		AcctName: "shankj3",
		Identifier: "apple prof",
		AppleSecrets: zipdata,
		AppleSecretsPassword: pw,
		SubType: pb.SubCredType_DEVPROFILE,
	}
	rc := &vcsRemoteConf{}
	gos := &guideOcelotServer{RemoteConfig:rc, RepoValidator: credentials.GetRepoValidator()}
	ctx := context.Background()
	_, err := gos.SetAppleCreds(ctx, creds)
	if err != nil {
		t.Error("creds are legit, remote conf is happy, this should work. err: " + err.Error())
	}
	creds.SubType = pb.SubCredType_SSHKEY
	_, err = gos.SetAppleCreds(ctx, creds)
	if err == nil {
		t.Error("wrong subtype, should fail validation")
	}
	if !strings.Contains(err.Error(), "Subtype must be of apple") {
		t.Error("should return subtype error")
	}
	creds.SubType = pb.SubCredType_DEVPROFILE
	rc.returnErr = true
	_, err = gos.SetAppleCreds(ctx, creds)
	if err == nil {
		t.Error("remote conf returns unknown error, this should be bubbled up")
	}
	rc.returnErr = false
	rc.validationErr = true
	creds.AppleSecrets = zipdata
	_, err = gos.SetAppleCreds(ctx, creds)
	if err == nil {
		t.Error("remote conf returns validation error, this should be bubbled up")
	}
	statErr, ok := status.FromError(err)
	if !ok {
		t.Error("should return admin grpc error")
	}
	if statErr.Code() != codes.InvalidArgument {
		t.Error("failed validation, should return grpc code of invalid argument")
	}
}

func TestGuideOcelotServer_GetAppleCreds(t *testing.T) {
	rc := &vcsRemoteConf{}
	gos := &guideOcelotServer{RemoteConfig:rc, RepoValidator: credentials.GetRepoValidator()}
	ctx := context.Background()
	creds, err := gos.GetAppleCreds(ctx, nil)
	if err != nil {
		t.Error(err)
	}
	if len(creds.AppleCreds) != 3 {
		t.Error("didn't get all apple creds? ")
	}
	rc.notFound = true
	_, err = gos.GetAppleCreds(ctx, nil)
	if err == nil {
		t.Error("should return not found error")
	}
	staterr, ok := status.FromError(err)
	if !ok {
		t.Error("should return grpc error")
	}
	if staterr.Code() != codes.NotFound {
		t.Error("should return code fo not found")
	}
	rc.notFound = false
	rc.empty = true
	_, err = gos.GetAppleCreds(ctx, nil)
	if err == nil {
		t.Error("should return not found error as returned apple list was empty")
	}
	staterr, ok = status.FromError(err)
	if !ok {
		t.Error("should return grpc error")
	}
	if staterr.Code() != codes.NotFound {
		t.Error("should return code fo not found")
	}
	rc.empty = false
	rc.returnErr = true
	_, err = gos.GetAppleCreds(ctx, nil)
	if err == nil {
		t.Error("should return  error as remote conf returned unidentifiable error")
	}
	staterr, ok = status.FromError(err)
	if !ok {
		t.Error("should return grpc error")
	}
	if staterr.Code() != codes.Internal {
		t.Error("should return code fo internal as error was unhandled")
	}
}

func TestGuideOcelotServer_GetAppleCred(t *testing.T) {
	rc := &vcsRemoteConf{}
	gos := &guideOcelotServer{RemoteConfig:rc, RepoValidator: credentials.GetRepoValidator()}
	ctx := context.Background()
	_, err := gos.GetAppleCred(ctx, &pb.AppleCreds{AcctName:"ay", Identifier:"id"})
	if err == nil {
		t.Error("subtype is not given, this request should return an error")
	}
	msg := "rpc error: code = InvalidArgument desc = subType, acctName, and identifier are required fields"
	if err.Error() != msg {
		t.Error(test.StrFormatErrors("err msg", msg, err.Error()))
	}
	rc.returnWrong = true
	_, err = gos.GetAppleCred(ctx, &pb.AppleCreds{AcctName:"ay", Identifier:"id", SubType:pb.SubCredType_DEVPROFILE})
	if err == nil {
		t.Error("returned wrong cred, shoudl fail")
	}

}

func TestGuideOcelotServer_UpdateAppleCreds(t *testing.T) {
	zipdata, pw := ioshelper.GetZipAndPw(t)
	creds := &pb.AppleCreds{
		AcctName: "shankj3",
		Identifier: "apple prof",
		AppleSecrets: zipdata,
		AppleSecretsPassword: pw,
		SubType: pb.SubCredType_DEVPROFILE,
	}
	rc := &vcsRemoteConf{}
	gos := &guideOcelotServer{RemoteConfig:rc}
	ctx := context.Background()
	_, err := gos.UpdateAppleCreds(ctx, creds)
	if err != nil {
		t.Error(err)
	}
	creds.AppleSecrets = []byte("derpyderpyderpy")
	_, err = gos.UpdateAppleCreds(ctx, creds)
	if err == nil {
		t.Error("bad apple secrets, should return an error")
	}
}

func TestGuideOcelotServer_AppleCredExists(t *testing.T) {
	rc := &vcsRemoteConf{}
	stor := &store{exists:true}
	gos := &guideOcelotServer{RemoteConfig:rc, Storage:stor}
	exists, err := gos.AppleCredExists(context.Background(), &pb.AppleCreds{})
	if err != nil {
		t.Error("should pass")
	}
	if !exists.Exists {
		t.Error("should return  existing")
	}
}

func TestGuideOcelotServer_SetNotifyCreds(t *testing.T) {
	rc := &vcsRemoteConf{}
	stor := &store{exists:true}
	gos := &guideOcelotServer{RemoteConfig:rc, Storage:stor}
	cred := &pb.NotifyCreds{
		AcctName:"shankj3",
		SubType:pb.SubCredType_SLACK,
		ClientSecret: "http://seceret.post",
		Identifier: "derp",
	}
	ctx := context.Background()
	_, err := gos.SetNotifyCreds(ctx, cred)
	if err != nil {
		t.Error(err)
	}
	cred.SubType = pb.SubCredType_SSHKEY
	_, err = gos.SetNotifyCreds(ctx, cred)
	if err == nil {
		t.Error("wrong subtype, this should return a validation error.")
	}
	cred.SubType = pb.SubCredType_SLACK
	cred.Identifier = "derp"
	rc.validationErr = true
	_, err = gos.SetNotifyCreds(ctx, cred)
	if err == nil {
		t.Error("remote conf returned a validation error, this should bubble up")
	}
	rc.validationErr = false
	rc.returnErr = true
	_, err = gos.SetNotifyCreds(ctx, cred)
	if err == nil {
		t.Error("remote conf returned an unknown error, this should bubble up")
	}
}

func TestGuideOcelotServer_NotifyCredExists(t *testing.T) {
	// tested extensively in checkAnyCredExists
	rc := &vcsRemoteConf{}
	stor := &store{exists:true}
	gos := &guideOcelotServer{RemoteConfig:rc, Storage:stor}
	exists, err := gos.NotifyCredExists(context.Background(), &pb.NotifyCreds{})
	if err != nil {
		t.Error("should pass")
	}
	if !exists.Exists {
		t.Error("should return  existing")
	}
}

func TestGuideOcelotServer_UpdateNotifyCreds(t *testing.T) {
	// tested extensively in updateAnyCred
	rc := &vcsRemoteConf{}
	gos := &guideOcelotServer{RemoteConfig:rc}
	ctx := context.Background()
	_, err := gos.UpdateNotifyCreds(ctx, &pb.NotifyCreds{})
	if err != nil {
		t.Error(err)
	}
}

func TestGuideOcelotServer_GetNotifyCred(t *testing.T) {
	rc := &vcsRemoteConf{}
	stor := &store{exists:true}
	gos := &guideOcelotServer{RemoteConfig:rc, Storage:stor}
	cred := &pb.NotifyCreds{
		Identifier: "slackky",
		SubType:pb.SubCredType_SLACK,
		AcctName: "accountname",
	}
	ctx := context.Background()
	_, err := gos.GetNotifyCred(ctx, cred)
	if err != nil {
		t.Error(err)
	}
	cred.Identifier = ""
	_, err = gos.GetNotifyCred(ctx, cred)
	if err == nil {
		t.Error("idenitifer is empty, should return error.")
	}
	cred.Identifier = "slaccky"
	rc.returnWrong = true
	_, err = gos.GetNotifyCred(ctx, cred)
	if err == nil {
		t.Error("returned wrong struct, should fail")
	}
}

func TestGuideOcelotServer_GetNotifyCreds(t *testing.T) {
	rc := &vcsRemoteConf{}
	stor := &store{exists:true}
	gos := &guideOcelotServer{RemoteConfig:rc, Storage:stor}
	ctx := context.Background()
	creds, err := gos.GetNotifyCreds(ctx, nil)
	if err != nil {
		t.Error(err)
	}
	if len(creds.Creds) != 3 {
		t.Error("wrong cred length")
	}
	rc.notFound = true
	_, err = gos.GetNotifyCreds(ctx, nil)
	if err == nil {
		t.Error("should return not found., er")
	}
	rc.notFound = false
	rc.empty = true
	_, err = gos.GetNotifyCreds(ctx, nil)
	if err == nil {
		t.Error("should return not found as cred list is empty")
	}
	rc.empty = false
	rc.returnErr = true
	_, err = gos.GetNotifyCreds(ctx, nil)
	if err == nil {
		t.Error("should return error")
	}
}

type vcsRemoteConf struct {
	credentials.CVRemoteConfig
	sshExists bool
	empty bool
	notFound bool
	returnErr bool
	validationErr bool
	returnWrong bool
	updated []pb.OcyCredder
}


var vcsCred = &pb.VCSCreds{
	ClientId: "1",
	ClientSecret: "2",
	Identifier: "sec",
	TokenURL: "http://token.url",
	SubType: pb.SubCredType_BITBUCKET,
	AcctName: "shankj3",
}

var repoCred = &pb.RepoCreds{
	AcctName: "shankj3",
	Password: "password",
	Username: "username",
	Identifier: "identitty",
	SubType: pb.SubCredType_SSHKEY,
}

var k8sCred = &pb.K8SCreds{
	AcctName:"shankj3",
	Identifier:"k8s",
	K8SContents: "hummunuauauaua",
	SubType: pb.SubCredType_KUBECONF,
}

var sshCred = &pb.SSHKeyWrapper{
	AcctName: "shankj3",
	Identifier:"ssh",
	PrivateKey: []byte("priv key"),
	SubType: pb.SubCredType_SSHKEY,
}
var appleCred = &pb.AppleCreds{
	AcctName: "shankj3",
	Identifier:"apple",
	AppleSecrets: []byte("secretzip"),
	AppleSecretsPassword: "pw",
	SubType: pb.SubCredType_DEVPROFILE,
}

var notifycred = &pb.NotifyCreds{
	AcctName:"shankj3",
	Identifier: "notify",
	SubType: pb.SubCredType_SLACK,
	ClientSecret: "secretive.",
}

func (r *vcsRemoteConf) GetCredsByType(store storage.CredTable, ctype pb.CredType, hideSecret bool) ([]pb.OcyCredder, error) {
	if r.empty {
		return []pb.OcyCredder{}, nil
	}
	if r.notFound {
		return nil, storage.CredNotFound("", ctype.String())
	}
	if r.returnErr {
		return nil, errors.New("returning an error")
	}
	switch ctype {
	case pb.CredType_VCS:
		return []pb.OcyCredder{vcsCred, vcsCred, vcsCred}, nil
	case pb.CredType_REPO:
		return []pb.OcyCredder{repoCred, repoCred, repoCred}, nil
	case pb.CredType_K8S:
		return []pb.OcyCredder{k8sCred, k8sCred, k8sCred}, nil
	case pb.CredType_SSH:
		return []pb.OcyCredder{sshCred, sshCred, sshCred}, nil
	case pb.CredType_APPLE:
		return []pb.OcyCredder{appleCred, appleCred, appleCred}, nil
	case pb.CredType_NOTIFIER:
		return []pb.OcyCredder{notifycred, notifycred, notifycred}, nil
	}
	return nil, errors.New("nope!!!")
}

func (r *vcsRemoteConf) CheckSSHKeyExists(path string) error {
	if !r.sshExists {
		return errors.New("no ssh key")
	}
	return nil
}

func (r *vcsRemoteConf) AddCreds(store storage.CredTable, anyCred pb.OcyCredder, overwriteOk bool) (err error) {
	if r.returnErr {
		return errors.New("this is an error")
	}
	if r.validationErr {
		return pb.Invalidate("this is invalid for AddCreds")
	}
	return nil
}

func (r *vcsRemoteConf) UpdateCreds(store storage.CredTable, anyCred pb.OcyCredder) (err error) {
	if r.returnErr {
		return errors.New("returning error from updateCreds")
	}
	if r.validationErr {
		return pb.Invalidate("not a valid cred for update")
	}
	r.updated = append(r.updated, anyCred)
	return nil
}


func (r *vcsRemoteConf) GetCred(store storage.CredTable, subCredType pb.SubCredType, identifier, accountName string, hideSecret bool) (pb.OcyCredder, error) {
	if r.returnErr {
		return nil, errors.New("this is an error")
	}
	if r.empty {
		return nil, nil
	}
	if r.notFound {
		return nil, storage.CredNotFound(accountName, subCredType.String())
	}
	if r.validationErr {
		return nil, pb.Invalidate("not valid, yo")
	}

	switch subCredType.Parent() {
	case pb.CredType_VCS:
		if r.returnWrong {
			return notifycred, nil
		}
		return &pb.VCSCreds{AcctName:accountName, Identifier:identifier, ClientSecret: "secret", ClientId: "xxxxx"}, nil
	case pb.CredType_REPO:
		if r.returnWrong {
			return notifycred, nil
		}
		return &pb.RepoCreds{AcctName:accountName, Identifier:identifier, Password:"secret", Username:"username", RepoUrl: "http://repo.url"}, nil
	case pb.CredType_K8S:
		if r.returnWrong {
			return notifycred, nil
		}
		return &pb.K8SCreds{AcctName:accountName, Identifier:identifier, SubType:subCredType, K8SContents:"hi!"}, nil
	case pb.CredType_SSH:
		if r.returnWrong {
			return notifycred, nil
		}
		return &pb.SSHKeyWrapper{AcctName:accountName, Identifier:identifier, SubType:subCredType, PrivateKey:[]byte("hi")}, nil
	case pb.CredType_APPLE:
		if r.returnWrong {
			return notifycred, nil
		}
		return &pb.AppleCreds{AcctName:accountName, Identifier:identifier, SubType:subCredType, AppleSecretsPassword:"pw", AppleSecrets: []byte("pw")}, nil
	case pb.CredType_NOTIFIER:
		if r.returnWrong {
			return appleCred, nil
		}
		return &pb.NotifyCreds{AcctName:accountName, Identifier:identifier, SubType:subCredType, ClientSecret: "http://slackurl.go"}, nil
	}
	return nil, errors.New("hehe not supported yet")

}

func (r *vcsRemoteConf) AddSSHKey(path string, sshKeyFile []byte) (err error) {
	if r.returnErr {
		return errors.New("error in AddSSHKey")
	}
	return nil
}

type store struct {
	exists bool
	returnErr bool
	storage.OcelotStorage
}

func (s *store) CredExists(credder pb.OcyCredder) (bool, error){
	if s.returnErr {
		return false, errors.New("erroring at cred existss check ")
	}
	return s.exists, nil
}
package secret

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/level11consulting/ocelot/secret/anycred"
	"github.com/level11consulting/ocelot/secret/appledev"
	"github.com/level11consulting/ocelot/secret/artifactrepo"
	"github.com/level11consulting/ocelot/secret/generic"
	"github.com/level11consulting/ocelot/secret/kubernetes"
	"github.com/level11consulting/ocelot/secret/notifier"
	"github.com/level11consulting/ocelot/secret/ssh"
	"github.com/level11consulting/ocelot/secret/vcs"
)

type SecretInterface interface {
	anycred.AnyCred
	appledev.AppleDevSecret
	artifactrepo.ArtifactRepoSecret
	generic.GenericSecret
	kubernetes.KubernetesSecret
	notifier.NotifierSecret
	ssh.SshSecret
	vcs.VcsSecret
	GetAllCreds(context.Context, *empty.Empty) (*pb.AllCredsWrapper, error)
}

type SecretInterfaceAPI struct {
	anycred.AnyCredAPI // This is a legacy catch-all hack
	appledev.AppleDevSecretAPI
	artifactrepo.ArtifactRepoSecretAPI
	generic.GenericSecretAPI
	kubernetes.KubernetesSecretAPI
	notifier.NotifierSecretAPI
	ssh.SshSecretAPI
	vcs.VcsSecretAPI
}

func (g *SecretInterfaceAPI) GetAllCreds(ctx context.Context, msg *empty.Empty) (*pb.AllCredsWrapper, error) {
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

package kubernetes

import (
	"fmt"
	"context"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/level11consulting/orbitalci/models/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/level11consulting/orbitalci/secret/legacy"
	"github.com/level11consulting/orbitalci/secret/anycred"
	"github.com/level11consulting/orbitalci/server/config"
	"github.com/level11consulting/orbitalci/storage"
)

type KubernetesSecret interface {
	GetK8SCred(context.Context, *pb.K8SCreds) (*pb.K8SCreds, error)
	UpdateK8SCreds(context.Context, *pb.K8SCreds) (*empty.Empty, error)
	K8SCredExists(context.Context, *pb.K8SCreds) (*pb.Exists, error)
	SetK8SCreds(context.Context, *pb.K8SCreds) (*empty.Empty, error)
	GetK8SCreds(context.Context, *empty.Empty) (*pb.K8SCredsWrapper, error)
	DeleteK8SCreds(context.Context, *pb.K8SCreds) (*empty.Empty, error)
}

type KubernetesSecretAPI struct {
	KubernetesSecret
	RemoteConfig   config.CVRemoteConfig
	Storage        storage.OcelotStorage
}

func (g *KubernetesSecretAPI) SetK8SCreds(ctx context.Context, creds *pb.K8SCreds) (*empty.Empty, error) {
	if creds.SubType.Parent() != pb.CredType_K8S {
		return nil, status.Error(codes.InvalidArgument, "Subtype must be of k8s type: "+strings.Join(pb.CredType_K8S.SubtypesString(), " | "))
	}
	// no validation necessary, its a file upload

	err := legacy.SetupRCCCredentials(g.RemoteConfig, g.Storage, creds)
	if err != nil {
		if _, ok := err.(*pb.ValidationErr); ok {
			return &empty.Empty{}, status.Error(codes.FailedPrecondition, "K8s Creds failed validation. Errors are: "+err.Error())
		}
		// todo: make this better error
		return &empty.Empty{}, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}

func (g *KubernetesSecretAPI) GetK8SCreds(ctx context.Context, empti *empty.Empty) (*pb.K8SCredsWrapper, error) {
	credWrapper := &pb.K8SCredsWrapper{}
	creds, err := g.RemoteConfig.GetCredsByType(g.Storage, pb.CredType_K8S, true)
	if err != nil {
		// todo: this needs to check for a not found error from storage as well
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

func (g *KubernetesSecretAPI) GetK8SCred(ctx context.Context, credentials *pb.K8SCreds) (*pb.K8SCreds, error) {
	creddy, err := g.GetAnyCred(credentials)
	if err != nil {
		return nil, err
	}
	repo, ok := creddy.(*pb.K8SCreds)
	if !ok {
		return nil, status.Error(codes.Internal, "Unable to cast as Kubernetes Creds")
	}
	return repo, nil
}

func (g *KubernetesSecretAPI) UpdateK8SCreds(ctx context.Context, creds *pb.K8SCreds) (*empty.Empty, error) {
	anyCredAPI := anycred.AnyCredAPI {
		Storage:        g.Storage,	
		RemoteConfig:   g.RemoteConfig,
	}

	return anyCredAPI.UpdateAnyCred(ctx, creds)
}

func (g *KubernetesSecretAPI) K8SCredExists(ctx context.Context, creds *pb.K8SCreds) (*pb.Exists, error) {
	anyCredAPI := anycred.AnyCredAPI {
		Storage:        g.Storage,	
		RemoteConfig:   g.RemoteConfig,
	}

	return anyCredAPI.CheckAnyCredExists(ctx, creds)
}

func (g *KubernetesSecretAPI) DeleteK8SCreds(ctx context.Context, creds *pb.K8SCreds) (*empty.Empty, error) {
	anyCredAPI := anycred.AnyCredAPI {
		Storage:        g.Storage,	
		RemoteConfig:   g.RemoteConfig,
	}

	return anyCredAPI.DeleteAnyCred(ctx, creds, pb.CredType_K8S)
}

func (g *KubernetesSecretAPI) GetAnyCred(credder pb.OcyCredder) (pb.OcyCredder, error) {
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

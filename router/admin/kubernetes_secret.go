package admin

import (
	"context"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/level11consulting/ocelot/models/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/level11consulting/ocelot/secret"
)

type KubernetesSecret interface {
	GetK8SCred(context.Context, *pb.K8SCreds) (*pb.K8SCreds, error)
	UpdateK8SCreds(context.Context, *pb.K8SCreds) (*empty.Empty, error)
	K8SCredExists(context.Context, *pb.K8SCreds) (*pb.Exists, error)
	SetK8SCreds(context.Context, *pb.K8SCreds) (*empty.Empty, error)
	GetK8SCreds(context.Context, *empty.Empty) (*pb.K8SCredsWrapper, error)
	DeleteK8SCreds(context.Context, *pb.K8SCreds) (*empty.Empty, error)
}

func (g *OcelotServerAPI) SetK8SCreds(ctx context.Context, creds *pb.K8SCreds) (*empty.Empty, error) {
	if creds.SubType.Parent() != pb.CredType_K8S {
		return nil, status.Error(codes.InvalidArgument, "Subtype must be of k8s type: "+strings.Join(pb.CredType_K8S.SubtypesString(), " | "))
	}
	// no validation necessary, its a file upload

	err := secret.SetupRCCCredentials(g.DeprecatedHandler.RemoteConfig, g.DeprecatedHandler.Storage, creds)
	if err != nil {
		if _, ok := err.(*pb.ValidationErr); ok {
			return &empty.Empty{}, status.Error(codes.FailedPrecondition, "K8s Creds failed validation. Errors are: "+err.Error())
		}
		// todo: make this better error
		return &empty.Empty{}, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}

func (g *OcelotServerAPI) GetK8SCreds(ctx context.Context, empti *empty.Empty) (*pb.K8SCredsWrapper, error) {
	credWrapper := &pb.K8SCredsWrapper{}
	creds, err := g.DeprecatedHandler.RemoteConfig.GetCredsByType(g.DeprecatedHandler.Storage, pb.CredType_K8S, true)
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

func (g *OcelotServerAPI) GetK8SCred(ctx context.Context, credentials *pb.K8SCreds) (*pb.K8SCreds, error) {
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

func (g *OcelotServerAPI) UpdateK8SCreds(ctx context.Context, creds *pb.K8SCreds) (*empty.Empty, error) {
	return g.UpdateAnyCred(ctx, creds)
}

func (g *OcelotServerAPI) K8SCredExists(ctx context.Context, creds *pb.K8SCreds) (*pb.Exists, error) {
	return g.CheckAnyCredExists(ctx, creds)
}

func (g *OcelotServerAPI) DeleteK8SCreds(ctx context.Context, creds *pb.K8SCreds) (*empty.Empty, error) {
	return g.DeleteAnyCred(ctx, creds, pb.CredType_K8S)
}

package models

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)
//type GuideOcelotClient interface {
//	GetCreds(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*CredWrapper, error)
//	SetCreds(ctx context.Context, in *Credentials, opts ...grpc.CallOption) (*empty.Empty, error)
//}

func NewFakeGuideOcelotClient() *fakeGuideOcelotClient {
	return &fakeGuideOcelotClient{creds: &CredWrapper{}}
}

type fakeGuideOcelotClient struct {
	creds *CredWrapper
}

func (f *fakeGuideOcelotClient) GetCreds(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*CredWrapper, error) {
	return f.creds, nil
}

func (f *fakeGuideOcelotClient) SetCreds(ctx context.Context, in *Credentials, opts ...grpc.CallOption) (*empty.Empty, error) {
	f.creds.Credentials = append(f.creds.Credentials, in)
	return &empty.Empty{}, nil
}

func CompareCredWrappers(credWrapA *CredWrapper, credWrapB *CredWrapper) bool {
	for ind, cred := range credWrapA.Credentials {
		credB := credWrapB.Credentials[ind]
		if cred.Type != credB.Type {
			return false
		}
		if cred.AcctName != credB.AcctName {
			return false
		}
		if cred.TokenURL != credB.TokenURL {
			return false
		}
		if cred.ClientSecret != credB.ClientSecret {
			return false
		}
		if cred.ClientId != credB.ClientId {
			return false
		}
	}
	return true
}
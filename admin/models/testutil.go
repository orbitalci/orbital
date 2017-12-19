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
	return &fakeGuideOcelotClient{creds: &CredWrapper{}, repoCreds: &RepoCredWrapper{}}
}

type fakeGuideOcelotClient struct {
	creds *CredWrapper
	repoCreds *RepoCredWrapper
}

func (f *fakeGuideOcelotClient) GetCreds(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*CredWrapper, error) {
	return f.creds, nil
}

func (f *fakeGuideOcelotClient) SetCreds(ctx context.Context, in *Credentials, opts ...grpc.CallOption) (*empty.Empty, error) {
	f.creds.Credentials = append(f.creds.Credentials, in)
	return &empty.Empty{}, nil
}

func (f *fakeGuideOcelotClient) GetRepoCreds(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*RepoCredWrapper, error) {
	return f.repoCreds, nil
}

func (f *fakeGuideOcelotClient) SetRepoCreds(ctx context.Context, in *RepoCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	f.repoCreds.Credentials = append(f.repoCreds.Credentials, in)
	return &empty.Empty{}, nil
}

func (f *fakeGuideOcelotClient) CheckConn(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func (f *fakeGuideOcelotClient) GetAllCreds(ctx context.Context, msg *empty.Empty, opts ...grpc.CallOption) (*AllCredsWrapper, error) {
	return &AllCredsWrapper{
		RepoCreds: f.repoCreds,
		AdminCreds: f.creds,
	}, nil
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

func CompareRepoCredWrappers(repoWrapA *RepoCredWrapper, repoWrapB *RepoCredWrapper) bool {
	for ind, cred := range repoWrapA.Credentials {
		credB := repoWrapB.Credentials[ind]
		if cred.Type != credB.Type {
			return false
		}
		if cred.Username != credB.Username {
			return false
		}
		if cred.AcctName != credB.AcctName {
			return false
		}
		if cred.Password != credB.Password {
			return false
		}
		if cred.RepoUrl != credB.RepoUrl {
			return false
		}
	}
	return true
}
package models

import (
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)
//type GuideOcelotClient interface {
//	GetVCSCreds(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*CredWrapper, error)
//	SetVCSCreds(ctx context.Context, in *Credentials, opts ...grpc.CallOption) (*empty.Empty, error)
//}

func NewFakeGuideOcelotClient() *fakeGuideOcelotClient {
	return &fakeGuideOcelotClient{creds: &CredWrapper{}, repoCreds: &RepoCredWrapper{}}
}

type fakeGuideOcelotClient struct {
	creds *CredWrapper
	repoCreds *RepoCredWrapper
}

func (f *fakeGuideOcelotClient) GetVCSCreds(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*CredWrapper, error) {
	return f.creds, nil
}

func (f *fakeGuideOcelotClient) SetVCSCreds(ctx context.Context, in *VCSCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	f.creds.Vcs = append(f.creds.Vcs, in)
	return &empty.Empty{}, nil
}

func (f *fakeGuideOcelotClient) GetRepoCreds(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*RepoCredWrapper, error) {
	return f.repoCreds, nil
}

func (f *fakeGuideOcelotClient) SetRepoCreds(ctx context.Context, in *RepoCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	f.repoCreds.Repo = append(f.repoCreds.Repo, in)
	return &empty.Empty{}, nil
}

func (f *fakeGuideOcelotClient) CheckConn(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func (f *fakeGuideOcelotClient) GetAllCreds(ctx context.Context, msg *empty.Empty, opts ...grpc.CallOption) (*AllCredsWrapper, error) {
	return &AllCredsWrapper{
		RepoCreds: f.repoCreds,
		VcsCreds: f.creds,
	}, nil
}

// todo: make this useful
func (f *fakeGuideOcelotClient) BuildRuntime(ctx context.Context, in *BuildQuery, opts ...grpc.CallOption) (*BuildRuntimeInfo, error) {
	return nil, nil
}

// todo: make this useful
func (f *fakeGuideOcelotClient) Logs(ctx context.Context, in *BuildQuery, opts ...grpc.CallOption) (GuideOcelot_LogsClient, error) {
	return nil, nil
}

func CompareCredWrappers(credWrapA *CredWrapper, credWrapB *CredWrapper) bool {
	for ind, cred := range credWrapA.Vcs {
		credB := credWrapB.Vcs[ind]
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
	for ind, cred := range repoWrapA.Repo {
		credB := repoWrapB.Repo[ind]
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

func CompareAllCredWrappers(allWrapA *AllCredsWrapper, allWrapB *AllCredsWrapper) bool {
	if repoMatches := CompareRepoCredWrappers(allWrapA.RepoCreds, allWrapB.RepoCreds); !repoMatches {
		return false
	}
	if vcsMatches := CompareCredWrappers(allWrapA.VcsCreds, allWrapB.VcsCreds); !vcsMatches {
		return false
	}
	return true
}
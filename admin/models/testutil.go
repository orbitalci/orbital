package models

import (
	"bitbucket.org/level11consulting/ocelot/werker/protobuf"
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"io"
	"github.com/golang/protobuf/ptypes/wrappers"
)
//type GuideOcelotClient interface {
//	GetVCSCreds(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*CredWrapper, error)
//	SetVCSCreds(ctx context.Context, in *Credentials, opts ...grpc.CallOption) (*empty.Empty, error)
//}

func NewFakeGuideOcelotClient(logLines []string) *fakeGuideOcelotClient {
	return &fakeGuideOcelotClient{creds: &CredWrapper{}, repoCreds: &RepoCredWrapper{}, logLines:logLines}
}

type fakeGuideOcelotClient struct {
	creds *CredWrapper
	repoCreds *RepoCredWrapper
	brInfo *Builds
	logLines []string
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

func (g *fakeGuideOcelotClient) StatusByHash(ctx context.Context, partialHash *wrappers.StringValue, opts ...grpc.CallOption) (*Status, error) {
	return &Status{}, nil
}


//todo: implement for testing
func (f *fakeGuideOcelotClient) LastFewSummaries(ctx context.Context, in *RepoAccount, opts ...grpc.CallOption) (*Summaries, error) {
	return &Summaries{}, nil
}


func (f *fakeGuideOcelotClient) BuildRuntime(ctx context.Context, in *BuildQuery, opts ...grpc.CallOption) (*Builds, error) {
	builds := &Builds{
		Builds: map[string]*BuildRuntimeInfo{},
	}
	//put your hash val and expected results here:
	switch in.Hash {
	case "testinghash":
		builds.Builds["abc"] = &BuildRuntimeInfo{
			Hash: "abc",
		}
		builds.Builds["def"] = &BuildRuntimeInfo{
			Hash: "def",
		}
	}

	return builds, nil
}

// todo: make this useful
func (f *fakeGuideOcelotClient) Logs(ctx context.Context, in *BuildQuery, opts ...grpc.CallOption) (GuideOcelot_LogsClient, error) {
	return NewFakeGuideOcelotLogsCli(f.logLines), nil
}

func NewFakeGuideOcelotLogsCli(lines []string) *fakeGuideOcelotLogsClient {
	return &fakeGuideOcelotLogsClient{outputLines: lines}
}

type fakeGuideOcelotLogsClient struct {
	index int
	outputLines []string
	grpc.ClientStream
}

func (c *fakeGuideOcelotLogsClient) CloseSend() error {
	return nil
}

func (c *fakeGuideOcelotLogsClient) Recv() (*LogResponse, error) {
	if c.index + 1 > len(c.outputLines) {
		return nil, io.EOF
	}
	resp := &LogResponse{OutputLine: c.outputLines[c.index]}
	c.index++
	return resp, nil
}

type testBuildClient struct {
	logLines []string
}

func (t *testBuildClient) BuildInfo(ctx context.Context, in *protobuf.Request, opts ...grpc.CallOption) (protobuf.Build_BuildInfoClient, error) {
	return protobuf.NewFakeBuildClient(t.logLines), nil
}

func NewTestBuildRuntime(done bool, ip string, grpcPort string, logLines []string) *testBuildRuntime{
	return &testBuildRuntime{
		Done: done,
		Ip: ip,
		GrpcPort: grpcPort,
		logLines: logLines,
	}
}

type testBuildRuntime struct {
	Done     bool
	Ip       string
	GrpcPort string
	logLines []string
	Hash	string
}

func (t *testBuildRuntime) GetDone() bool {
	return t.Done
}

func (t *testBuildRuntime) GetIp() string {
	return t.Ip
}

func (t *testBuildRuntime) GetGrpcPort() string {
	return t.GrpcPort
}

func (t *testBuildRuntime) GetHash() string {
	return t.Hash
}

func (t *testBuildRuntime) CreateBuildClient(opts []grpc.DialOption) (protobuf.BuildClient, error) {
	return &testBuildClient{logLines: t.logLines}, nil
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
		for name, url := range cred.RepoUrl {
			if m, ok := credB.RepoUrl[name]; !ok {
				return false
			} else if m != url {
				return false
			}
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
package testutil

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"github.com/level11consulting/orbitalci/models/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"io"
)

//type GuideOcelotClient interface {
//	GetVCSCreds(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*CredWrapper, error)
//	SetVCSCreds(ctx context.Context, in *Credentials, opts ...grpc.CallOption) (*empty.Empty, error)
//}

func NewFakeGuideOcelotClient(logLines []string) *FakeGuideOcelotClient {
	return &FakeGuideOcelotClient{creds: &pb.CredWrapper{}, repoCreds: &pb.RepoCredWrapper{}, logLines: logLines}
}

type FakeGuideOcelotClient struct {
	pb.GuideOcelotClient
	creds        *pb.CredWrapper
	repoCreds    *pb.RepoCredWrapper
	k8sCreds     *pb.K8SCredsWrapper
	Generics     *pb.GenericWrap
	ReturnError  bool
	brInfo       *pb.Builds
	Exists       bool
	logLines     []string
	NotConnected bool
}

func (f *FakeGuideOcelotClient) GetTrackedRepos(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.AcctRepos, error) {
	return nil, nil
}

func (f *FakeGuideOcelotClient) GetVCSCred(ctx context.Context, in *pb.VCSCreds, opts ...grpc.CallOption) (*pb.VCSCreds, error) {
	return nil, nil
}

func (f *FakeGuideOcelotClient) GetRepoCred(ctx context.Context, in *pb.RepoCreds, opts ...grpc.CallOption) (*pb.RepoCreds, error) {
	return nil, nil
}

func (f *FakeGuideOcelotClient) GetK8SCred(ctx context.Context, in *pb.K8SCreds, opts ...grpc.CallOption) (*pb.K8SCreds, error) {
	return nil, nil
}

func (f *FakeGuideOcelotClient) GetSSHCred(ctx context.Context, in *pb.SSHKeyWrapper, opts ...grpc.CallOption) (*pb.SSHKeyWrapper, error) {
	return nil, nil
}

func (f *FakeGuideOcelotClient) SetAppleCreds(ctx context.Context, in *pb.AppleCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, nil
}

func (f *FakeGuideOcelotClient) GetAppleCreds(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.AppleCredsWrapper, error) {
	return nil, nil
}

func (f *FakeGuideOcelotClient) GetAppleCred(ctx context.Context, in *pb.AppleCreds, opts ...grpc.CallOption) (*pb.AppleCreds, error) {
	return nil, nil
}

func (f *FakeGuideOcelotClient) UpdateAppleCreds(ctx context.Context, in *pb.AppleCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, nil
}
func (f *FakeGuideOcelotClient) AppleCredExists(ctx context.Context, in *pb.AppleCreds, opts ...grpc.CallOption) (*pb.Exists, error) {
	return nil, nil
}

func (f *FakeGuideOcelotClient) GetVCSCreds(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.CredWrapper, error) {
	return f.creds, nil
}

func (f *FakeGuideOcelotClient) SetVCSCreds(ctx context.Context, in *pb.VCSCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	in.SshFileLoc = "THIS IS A TEST"
	f.creds.Vcs = append(f.creds.Vcs, in)
	return &empty.Empty{}, nil
}

func (f *FakeGuideOcelotClient) UpdateVCSCreds(ctx context.Context, in *pb.VCSCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	for _, cred := range f.creds.Vcs {
		if cred.Identifier == in.Identifier && cred.AcctName == in.AcctName && cred.SubType == in.SubType {
			fmt.Println("setting cred")
			cred = in
		}
	}
	return nil, nil
}

func (f *FakeGuideOcelotClient) VCSCredExists(ctx context.Context, in *pb.VCSCreds, opts ...grpc.CallOption) (*pb.Exists, error) {
	for _, cred := range f.creds.Vcs {
		if cred.Identifier == in.Identifier && cred.AcctName == in.AcctName && cred.SubType == in.SubType {
			return &pb.Exists{Exists: true}, nil
		}
	}
	return &pb.Exists{Exists: false}, nil
}

func (f *FakeGuideOcelotClient) SetK8SCreds(ctx context.Context, in *pb.K8SCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	f.k8sCreds.K8SCreds = append(f.k8sCreds.K8SCreds, in)
	return &empty.Empty{}, nil
}

func (f *FakeGuideOcelotClient) UpdateK8SCreds(ctx context.Context, in *pb.K8SCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	for _, cred := range f.k8sCreds.K8SCreds {
		if cred.Identifier == in.Identifier && cred.AcctName == in.AcctName && cred.SubType == in.SubType {
			fmt.Println("setting cred")
			cred = in
		}
	}
	return nil, nil
}

func (f *FakeGuideOcelotClient) K8SCredExists(ctx context.Context, in *pb.K8SCreds, opts ...grpc.CallOption) (*pb.Exists, error) {
	for _, cred := range f.k8sCreds.K8SCreds {
		if cred.Identifier == in.Identifier && cred.AcctName == in.AcctName && cred.SubType == in.SubType {
			return &pb.Exists{Exists: true}, nil
		}
	}
	return &pb.Exists{Exists: false}, nil
}

func (f *FakeGuideOcelotClient) GetK8SCreds(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.K8SCredsWrapper, error) {
	return f.k8sCreds, nil
}

func (f *FakeGuideOcelotClient) SetSSHCreds(ctx context.Context, in *pb.SSHKeyWrapper, opts ...grpc.CallOption) (*empty.Empty, error) {
	//f.k8sCreds.K8SCreds = append(f.k8sCreds.K8SCreds, in)
	return &empty.Empty{}, nil
}

func (f *FakeGuideOcelotClient) UpdateSSHCreds(ctx context.Context, in *pb.SSHKeyWrapper, opts ...grpc.CallOption) (*empty.Empty, error) {
	//for _, cred := range f.k8sCreds.K8SCreds {
	//if cred.Identifier == in.Identifier && cred.AcctName == in.AcctName && cred.SubType == in.SubType {
	//	fmt.Println("setting cred")
	//	cred = in
	//}
	//}
	return nil, nil
}

func (f *FakeGuideOcelotClient) SSHCredExists(ctx context.Context, in *pb.SSHKeyWrapper, opts ...grpc.CallOption) (*pb.Exists, error) {
	//for _, cred := range f.k8sCreds.K8SCreds {
	//	if cred.Identifier == in.Identifier && cred.AcctName == in.AcctName && cred.SubType == in.SubType {
	//		return &pb.Exists{Exists:true}, nil
	//	}
	//}
	return &pb.Exists{Exists: false}, nil
}

func (f *FakeGuideOcelotClient) GetSSHCreds(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.SSHWrap, error) {
	return nil, nil
}
func (f *FakeGuideOcelotClient) SetNotifyCreds(ctx context.Context, in *pb.NotifyCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, nil
}

func (f *FakeGuideOcelotClient) GetNotifyCred(ctx context.Context, in *pb.NotifyCreds, opts ...grpc.CallOption) (*pb.NotifyCreds, error) {
	return nil, nil
}

func (f *FakeGuideOcelotClient) GetNotifyCreds(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.NotifyWrap, error) {
	return nil, nil
}

func (f *FakeGuideOcelotClient) UpdateNotifyCreds(ctx context.Context, in *pb.NotifyCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	return nil, nil
}

func (f *FakeGuideOcelotClient) NotifyCredExists(ctx context.Context, in *pb.NotifyCreds, opts ...grpc.CallOption) (*pb.Exists, error) {
	return nil, nil
}

func (f *FakeGuideOcelotClient) WatchRepo(ctx context.Context, in *pb.RepoAccount, opts ...grpc.CallOption) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func (f *FakeGuideOcelotClient) GetRepoCreds(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.RepoCredWrapper, error) {
	return f.repoCreds, nil
}

func (f *FakeGuideOcelotClient) SetRepoCreds(ctx context.Context, in *pb.RepoCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	f.repoCreds.Repo = append(f.repoCreds.Repo, in)
	in.SubType = pb.SubCredType_NEXUS
	return &empty.Empty{}, nil
}

func (f *FakeGuideOcelotClient) UpdateRepoCreds(ctx context.Context, in *pb.RepoCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	for _, cred := range f.repoCreds.Repo {
		if cred.Identifier == in.Identifier && cred.AcctName == in.AcctName && cred.SubType == in.SubType {
			fmt.Println("setting cred")
			cred = in
		}
	}
	return nil, nil
}

func (f *FakeGuideOcelotClient) RepoCredExists(ctx context.Context, in *pb.RepoCreds, opts ...grpc.CallOption) (*pb.Exists, error) {
	for _, cred := range f.repoCreds.Repo {
		if cred.Identifier == in.Identifier && cred.AcctName == in.AcctName && cred.SubType == in.SubType {
			return &pb.Exists{Exists: true}, nil
		}
	}
	return &pb.Exists{Exists: false}, nil
}

func (f *FakeGuideOcelotClient) CheckConn(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error) {
	if f.NotConnected {
		return nil, status.Error(codes.Internal, "connection failed :( ")
	}
	return &empty.Empty{}, nil
}

func (f *FakeGuideOcelotClient) GetAllCreds(ctx context.Context, msg *empty.Empty, opts ...grpc.CallOption) (*pb.AllCredsWrapper, error) {
	return &pb.AllCredsWrapper{
		RepoCreds: f.repoCreds,
		VcsCreds:  f.creds,
	}, nil
}

func (g *FakeGuideOcelotClient) GetStatus(ctx context.Context, query *pb.StatusQuery, opts ...grpc.CallOption) (*pb.Status, error) {
	return &pb.Status{}, nil
}

func (f *FakeGuideOcelotClient) SetVCSPrivateKey(ctx context.Context, in *pb.SSHKeyWrapper, opts ...grpc.CallOption) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

//todo: implement for testing
func (f *FakeGuideOcelotClient) LastFewSummaries(ctx context.Context, in *pb.RepoAccount, opts ...grpc.CallOption) (*pb.Summaries, error) {
	return &pb.Summaries{}, nil
}

func (f *FakeGuideOcelotClient) BuildRuntime(ctx context.Context, in *pb.BuildQuery, opts ...grpc.CallOption) (*pb.Builds, error) {
	builds := &pb.Builds{
		Builds: map[string]*pb.BuildRuntimeInfo{},
	}
	//put your hash val and expected results here:
	switch in.Hash {
	case "testinghash":
		builds.Builds["abc"] = &pb.BuildRuntimeInfo{
			Hash: "abc",
		}
		builds.Builds["def"] = &pb.BuildRuntimeInfo{
			Hash: "def",
		}
	}

	return builds, nil
}

func (f *FakeGuideOcelotClient) FindWerker(ctx context.Context, in *pb.BuildReq, opts ...grpc.CallOption) (*pb.BuildRuntimeInfo, error) {
	var build = &pb.BuildRuntimeInfo{
		Hash: "abc",
	}
	return build, nil
}

// todo: make this useful
func (f *FakeGuideOcelotClient) Logs(ctx context.Context, in *pb.BuildQuery, opts ...grpc.CallOption) (pb.GuideOcelot_LogsClient, error) {
	return NewFakeGuideOcelotLogsCli(f.logLines), nil
}

func (f *FakeGuideOcelotClient) BuildRepoAndHash(ctx context.Context, in *pb.BuildReq, opts ...grpc.CallOption) (pb.GuideOcelot_BuildRepoAndHashClient, error) {
	return nil, nil
}

func (f *FakeGuideOcelotClient) PollRepo(ctx context.Context, poll *pb.PollRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func (f *FakeGuideOcelotClient) DeletePollRepo(ctx context.Context, poll *pb.PollRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func (f *FakeGuideOcelotClient) ListPolledRepos(ctx context.Context, empti *empty.Empty, opts ...grpc.CallOption) (*pb.Polls, error) {
	return &pb.Polls{}, nil
}

func (f *FakeGuideOcelotClient) UpdateGenericCreds(ctx context.Context, in *pb.GenericCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	if f.ReturnError {
		return nil, errors.New("returning an erro")
	}
	f.Generics.Creds = append(f.Generics.Creds, in)
	return nil, nil
}

func (f *FakeGuideOcelotClient) GetGenericCreds(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.GenericWrap, error) {
	if f.ReturnError {
		return nil, errors.New("returning an erro get ")
	}
	return f.Generics, nil
}

func (f *FakeGuideOcelotClient) SetGenericCreds(ctx context.Context, in *pb.GenericCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	if f.ReturnError {
		return nil, errors.New("returning an erro on set")
	}
	f.Generics.Creds = append(f.Generics.Creds, in)
	return nil, nil
}

func (f *FakeGuideOcelotClient) GenericCredExists(ctx context.Context, in *pb.GenericCreds, opts ...grpc.CallOption) (*pb.Exists, error) {
	if f.ReturnError {
		return nil, errors.New("returning an erro on set")
	}
	return &pb.Exists{Exists: f.Exists}, nil
}

func NewFakeGuideOcelotLogsCli(lines []string) *fakeGuideOcelotLogsClient {
	return &fakeGuideOcelotLogsClient{outputLines: lines}
}

type fakeGuideOcelotLogsClient struct {
	index       int
	outputLines []string
	grpc.ClientStream
}

func (c *fakeGuideOcelotLogsClient) CloseSend() error {
	return nil
}

func (c *fakeGuideOcelotLogsClient) Recv() (*pb.LineResponse, error) {
	if c.index+1 > len(c.outputLines) {
		return nil, io.EOF
	}
	resp := &pb.LineResponse{OutputLine: c.outputLines[c.index]}
	c.index++
	return resp, nil
}

type testBuildClient struct {
	logLines []string
}

//type BuildClient interface {
//BuildInfo(ctx context.Context, in *Request, opts ...grpc.CallOption) (Build_BuildInfoClient, error)
//KillHash(ctx context.Context, in *Request, opts ...grpc.CallOption) (Build_KillHashClient, error)
//}

func (t *testBuildClient) BuildInfo(ctx context.Context, in *pb.Request, opts ...grpc.CallOption) (pb.Build_BuildInfoClient, error) {
	return NewFakeBuildClient(t.logLines), nil
}

func (t *testBuildClient) KillHash(ctx context.Context, in *pb.Request, opts ...grpc.CallOption) (pb.Build_KillHashClient, error) {
	return NewFakeBuildClient(t.logLines), nil
}

func NewTestBuildRuntime(done bool, ip string, grpcPort string, logLines []string) *testBuildRuntime {
	return &testBuildRuntime{
		Done:     done,
		Ip:       ip,
		GrpcPort: grpcPort,
		logLines: logLines,
	}
}

type testBuildRuntime struct {
	Done     bool
	Ip       string
	GrpcPort string
	logLines []string
	Hash     string
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

func (t *testBuildRuntime) CreateBuildClient() (pb.BuildClient, error) {
	return &testBuildClient{logLines: t.logLines}, nil
}

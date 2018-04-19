package testutil


import (
	"bitbucket.org/level11consulting/ocelot/models/pb"
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"fmt"
	"io"
)
//type GuideOcelotClient interface {
//	GetVCSCreds(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*CredWrapper, error)
//	SetVCSCreds(ctx context.Context, in *Credentials, opts ...grpc.CallOption) (*empty.Empty, error)
//}

func NewFakeGuideOcelotClient(logLines []string) *fakeGuideOcelotClient {
	return &fakeGuideOcelotClient{creds: &pb.CredWrapper{}, repoCreds: &pb.RepoCredWrapper{}, logLines:logLines}
}

type fakeGuideOcelotClient struct {
	creds *pb.CredWrapper
	repoCreds *pb.RepoCredWrapper
	k8sCreds *pb.K8SCredsWrapper
	brInfo *pb.Builds
	logLines []string
}

func (f *fakeGuideOcelotClient) GetVCSCreds(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.CredWrapper, error) {
	return f.creds, nil
}

func (f *fakeGuideOcelotClient) SetVCSCreds(ctx context.Context, in *pb.VCSCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	in.SshFileLoc = "THIS IS A TEST"
	f.creds.Vcs = append(f.creds.Vcs, in)
	return &empty.Empty{}, nil
}

func (f *fakeGuideOcelotClient) UpdateVCSCreds(ctx context.Context, in *pb.VCSCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	for _, cred := range f.creds.Vcs {
		if cred.Identifier == in.Identifier && cred.AcctName == in.AcctName && cred.SubType == in.SubType {
			fmt.Println("setting cred")
			cred = in
		}
	}
	return nil, nil
}

func (f *fakeGuideOcelotClient) VCSCredExists(ctx context.Context, in *pb.VCSCreds, opts ...grpc.CallOption) (*pb.Exists, error) {
	for _, cred := range f.creds.Vcs {
		if cred.Identifier == in.Identifier && cred.AcctName == in.AcctName && cred.SubType == in.SubType {
			return &pb.Exists{Exists:true}, nil
		}
	}
	return &pb.Exists{Exists:false}, nil
}

func (f *fakeGuideOcelotClient) SetK8SCreds(ctx context.Context, in *pb.K8SCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	f.k8sCreds.K8SCreds = append(f.k8sCreds.K8SCreds, in)
	return &empty.Empty{}, nil
}

func (f *fakeGuideOcelotClient) UpdateK8SCreds(ctx context.Context, in *pb.K8SCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	for _, cred := range f.k8sCreds.K8SCreds {
		if cred.Identifier == in.Identifier && cred.AcctName == in.AcctName && cred.SubType == in.SubType {
			fmt.Println("setting cred")
			cred = in
		}
	}
	return nil, nil
}

func (f *fakeGuideOcelotClient) K8SCredExists (ctx context.Context, in *pb.K8SCreds, opts ...grpc.CallOption) (*pb.Exists, error) {
	for _, cred := range f.k8sCreds.K8SCreds {
		if cred.Identifier == in.Identifier && cred.AcctName == in.AcctName && cred.SubType == in.SubType {
			return &pb.Exists{Exists:true}, nil
		}
	}
	return &pb.Exists{Exists:false}, nil
}


func (f *fakeGuideOcelotClient) GetK8SCreds(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.K8SCredsWrapper, error) {
	return f.k8sCreds, nil
}


func (f *fakeGuideOcelotClient) WatchRepo(ctx context.Context, in *pb.RepoAccount, opts ...grpc.CallOption) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func (f *fakeGuideOcelotClient) GetRepoCreds(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*pb.RepoCredWrapper, error) {
	return f.repoCreds, nil
}

func (f *fakeGuideOcelotClient) SetRepoCreds(ctx context.Context, in *pb.RepoCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	f.repoCreds.Repo = append(f.repoCreds.Repo, in)
	in.SubType = pb.SubCredType_NEXUS
	return &empty.Empty{}, nil
}

func (f *fakeGuideOcelotClient) UpdateRepoCreds(ctx context.Context, in *pb.RepoCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	for _, cred := range f.repoCreds.Repo {
		if cred.Identifier == in.Identifier && cred.AcctName == in.AcctName && cred.SubType == in.SubType {
			fmt.Println("setting cred")
			cred = in
		}
	}
	return nil, nil
}

func (f *fakeGuideOcelotClient) RepoCredExists(ctx context.Context, in *pb.RepoCreds, opts ...grpc.CallOption) (*pb.Exists, error) {
	for _, cred := range f.repoCreds.Repo {
		if cred.Identifier == in.Identifier && cred.AcctName == in.AcctName && cred.SubType == in.SubType {
			return &pb.Exists{Exists:true}, nil
		}
	}
	return &pb.Exists{Exists:false}, nil
}

func (f *fakeGuideOcelotClient) CheckConn(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func (f *fakeGuideOcelotClient) GetAllCreds(ctx context.Context, msg *empty.Empty, opts ...grpc.CallOption) (*pb.AllCredsWrapper, error) {
	return &pb.AllCredsWrapper{
		RepoCreds: f.repoCreds,
		VcsCreds: f.creds,
	}, nil
}

func (g *fakeGuideOcelotClient) GetStatus(ctx context.Context, query *pb.StatusQuery, opts ...grpc.CallOption) (*pb.Status, error) {
	return &pb.Status{}, nil
}


func (f *fakeGuideOcelotClient)	SetVCSPrivateKey(ctx context.Context, in *pb.SSHKeyWrapper, opts ...grpc.CallOption) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}



//todo: implement for testing
func (f *fakeGuideOcelotClient) LastFewSummaries(ctx context.Context, in *pb.RepoAccount, opts ...grpc.CallOption) (*pb.Summaries, error) {
	return &pb.Summaries{}, nil
}


func (f *fakeGuideOcelotClient) BuildRuntime(ctx context.Context, in *pb.BuildQuery, opts ...grpc.CallOption) (*pb.Builds, error) {
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

func (f *fakeGuideOcelotClient) FindWerker(ctx context.Context, in *pb.BuildReq, opts ...grpc.CallOption) (*pb.BuildRuntimeInfo, error) {
	var build = &pb.BuildRuntimeInfo{
		Hash: "abc",
	}
	return build, nil
}

// todo: make this useful
func (f *fakeGuideOcelotClient) Logs(ctx context.Context, in *pb.BuildQuery, opts ...grpc.CallOption) (pb.GuideOcelot_LogsClient, error) {
	return NewFakeGuideOcelotLogsCli(f.logLines), nil
}

func (f *fakeGuideOcelotClient) BuildRepoAndHash(ctx context.Context, in *pb.BuildReq, opts ...grpc.CallOption) (pb.GuideOcelot_BuildRepoAndHashClient, error) {
	return nil, nil
}

func (f *fakeGuideOcelotClient) PollRepo(ctx context.Context, poll *pb.PollRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func (f *fakeGuideOcelotClient) DeletePollRepo(ctx context.Context, poll *pb.PollRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func (f *fakeGuideOcelotClient) ListPolledRepos(ctx context.Context, empti *empty.Empty, opts ...grpc.CallOption) (*pb.Polls, error) {
	return &pb.Polls{}, nil
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

func (c *fakeGuideOcelotLogsClient) Recv() (*pb.LineResponse, error) {
	if c.index + 1 > len(c.outputLines) {
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

func (t *testBuildRuntime) CreateBuildClient() (pb.BuildClient, error) {
	return &testBuildClient{logLines: t.logLines}, nil
}

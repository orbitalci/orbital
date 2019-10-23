package delete

import (
	"bytes"
	"context"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mitchellh/cli"
	"github.com/shankj3/go-til/test"
	"github.com/level11consulting/orbitalci/client/commandhelper"
	"github.com/level11consulting/orbitalci/models/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestCmd_Run(t *testing.T) {
	ui := cli.NewMockUi()
	clie := &fakeCli{}
	cmdd := &cmd{UI: ui, config: &commandhelper.ClientConfig{Client: clie}}
	cmdd.init()
	ui.InputReader = bytes.NewReader([]byte("YES"))
	cmdd.credType = pb.CredType_NOTIFIER
	exitcode := cmdd.Run([]string{"-acct=12", "-identifier=123"})
	if exitcode != 0 {
		t.Error("wrong exit code")
	}
	expectedOutput := "Are you sure that you want to delete this credential? This action is irreversible! Type YES if you mean it.Successfully deleted SLACK credential under account 12 and identifier 123\n"
	out := ui.OutputWriter.String()
	if expectedOutput != out {
		t.Error(test.StrFormatErrors("output", expectedOutput, out))
	}
	if len(clie.notifiers) != 1 {
		t.Error("did not delete correct notify type")
	}

	ui.OutputWriter.Reset()

	cmdd.credType = pb.CredType_GENERIC
	ui.InputReader = bytes.NewBuffer([]byte("es"))
	exitcode = cmdd.Run([]string{"-acct=12", "-identifier=123", "-subtype=env"})
	if exitcode != 0 {
		t.Error("After cancelling out of delete confirmation, should return 0 exit code")
	}
	expectedOutput = "Are you sure that you want to delete this credential? This action is irreversible! Type YES if you mean it.YES not entered, not deleting credential... \n"
	if expectedOutput != ui.OutputWriter.String() {
		t.Error(test.StrFormatErrors("output", expectedOutput, ui.OutputWriter.String()))
	}
	ui.OutputWriter.Reset()

	ui.InputReader = bytes.NewBuffer([]byte("YES"))
	exitcode = cmdd.Run([]string{"-acct=12", "-identifier=123"})
	if exitcode != 0 {
		t.Error("should return exit code of 0")
	}
	if len(clie.generics) != 1 {
		t.Error("did not call correct function")
	}

	// repo has multiple subtypes, should return error saying that you have to specify subtype
	cmdd.credType = pb.CredType_REPO
	ui.InputReader = bytes.NewBuffer([]byte("YES"))
	exitcode = cmdd.Run([]string{"-acct=12", "-identifier=123"})
	if exitcode != 1 {
		t.Error("should return 1 exit code as bad input ")
	}
	expectedOutput = "Not the correct subtype for this credential (REPO). Please pick one of NEXUS|MAVEN|DOCKER|MINIO\n"
	live := ui.ErrorWriter.String()
	if expectedOutput != live {
		t.Error(test.StrFormatErrors("error output", expectedOutput, live))
	}
	// test w/ subtype
	ui.ErrorWriter.Reset()

	exitcode = cmdd.Run([]string{"-acct=12", "-identifier=123", "-subtype=MINIO"})
	if exitcode != 0 {
		t.Error("subtype was specified, this should pass")
	}
	if len(clie.repos) != 1 {
		t.Error("did not call correct delete function")
	}
	// test w/ bad subtype
	ui.ErrorWriter.Reset()
	exitcode = cmdd.Run([]string{"-acct=12", "-identifier=123", "-subtype=mini"})
	if exitcode != 1 {
		t.Error("should return error as bad subtype")
	}
	expectedOutput = `Not found in the subtype map, please check spelling. 
Yours: mini
Available: NEXUS|MAVEN|DOCKER|MINIO
`
	if expectedOutput != ui.ErrorWriter.String() {
		t.Error(test.StrFormatErrors("error output", expectedOutput, ui.ErrorWriter.String()))
	}

	clie.fail = true
	ui.ErrorWriter.Reset()
	ui.InputReader = bytes.NewBuffer([]byte("YES"))
	exitcode = cmdd.Run([]string{"-acct=12", "-identifier=123", "-subtype=MINIO"})
	if exitcode != 1 {
		t.Error("client failed to delete credential, this should fail")
	}
	expectedOutput = "An error (MINIO) has occured: failez cred\n"
	if expectedOutput != ui.ErrorWriter.String() {
		t.Error(test.StrFormatErrors("error output", expectedOutput, ui.ErrorWriter.String()))
	}

	clie.fail = false
	cmdd.account = ""
	cmdd.identifier = ""
	cmdd.init()
	ui.ErrorWriter.Reset()
	exitcode = cmdd.Run([]string{})
	if ui.ErrorWriter.String() != "-acct was not provided\n" {
		t.Error(test.StrFormatErrors("no acct error", "-acct was not provided\n", ui.ErrorWriter.String()))
	}

	ui.ErrorWriter.Reset()
	exitcode = cmdd.Run([]string{"-acct=here"})
	if ui.ErrorWriter.String() != "-identifier required\n" {
		t.Error(test.StrFormatErrors("no identifier error", "-identifier required\n", ui.ErrorWriter.String()))
	}
	//

	clie.fail = true
	clie.fail404 = true
	cmdd.init()
	ui.ErrorWriter.Reset()
	cmdd.credType = pb.CredType_K8S
	ui.InputReader = bytes.NewBuffer([]byte("YES"))
	exitcode = cmdd.Run([]string{"-acct=here", "-identifier=there"})
	if exitcode != 1 {
		t.Error("client failed, should return 1")
	}
	if ui.ErrorWriter.String() != "Credential not found\n" {
		t.Error("should return not found")
		t.Log(ui.ErrorWriter.String())
	}
}

type fakeCli struct {
	pb.GuideOcelotClient
	failConn  bool
	fail      bool
	fail404   bool
	generics  []*pb.GenericCreds
	notifiers []*pb.NotifyCreds
	sshers    []*pb.SSHKeyWrapper
	repos     []*pb.RepoCreds
	k8s       []*pb.K8SCreds
}

func (f *fakeCli) CheckConn(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error) {
	if f.failConn {
		return nil, status.Error(codes.Internal, "failez")
	}
	return nil, nil
}

func (f *fakeCli) DeleteGenericCreds(ctx context.Context, in *pb.GenericCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	if f.fail {
		return nil, status.Error(codes.Internal, "failez cred")
	}
	f.generics = append(f.generics, in)
	return nil, nil
}

func (f *fakeCli) DeleteK8SCreds(ctx context.Context, in *pb.K8SCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	if f.fail {
		code := codes.Internal
		if f.fail404 {
			code = codes.NotFound
		}
		return nil, status.Error(code, "failez cred")
	}
	f.k8s = append(f.k8s, in)
	return nil, nil
}

func (f *fakeCli) DeleteNotifyCreds(ctx context.Context, in *pb.NotifyCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	if f.fail {
		return nil, status.Error(codes.Internal, "failez cred")
	}
	f.notifiers = append(f.notifiers, in)
	return nil, nil
}

func (f *fakeCli) DeleteSSHCreds(ctx context.Context, in *pb.SSHKeyWrapper, opts ...grpc.CallOption) (*empty.Empty, error) {
	if f.fail {
		return nil, status.Error(codes.Internal, "failez cred")
	}
	f.sshers = append(f.sshers, in)
	return nil, nil
}

func (f *fakeCli) DeleteRepoCreds(ctx context.Context, in *pb.RepoCreds, opts ...grpc.CallOption) (*empty.Empty, error) {
	if f.fail {
		return nil, status.Error(codes.Internal, "failez cred")
	}
	f.repos = append(f.repos, in)
	return nil, nil
}

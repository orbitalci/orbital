package build

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mitchellh/cli"
	"github.com/shankj3/ocelot/client/commandhelper"
	"github.com/shankj3/ocelot/models/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeCli struct {
	pb.GuideOcelotClient
	stream *fakeStream
	returnErr bool
	buildReq *pb.BuildReq
}

func (f *fakeCli) BuildRepoAndHash(ctx context.Context, buildReq *pb.BuildReq, opts ...grpc.CallOption) (pb.GuideOcelot_BuildRepoAndHashClient, error){
	f.buildReq = buildReq
	if f.returnErr {
		return nil, status.Error(codes.Internal, "this is an error.")
	}
	return f.stream, nil
}

func (f *fakeCli) CheckConn(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

type fakeStream struct {
	countBeforeEOF int
	count 		   int
	fail           bool
	pb.GuideOcelot_BuildRepoAndHashClient
}

func (f *fakeStream) CloseSend() error {
	return nil
}

func (f *fakeStream) RecvMsg(intr interface{}) error {
	if f.fail {
		return status.Error(codes.Internal, "this is an error")
	}
	if f.count == f.countBeforeEOF {
		return io.EOF
	}
	msg := intr.(*pb.Response)
	msg.OutputLine = fmt.Sprintf("line number %d of %d", f.count, f.countBeforeEOF)
	f.count++
	return nil
}

func TestCmd_Run(t *testing.T) {
	stream := &fakeStream{countBeforeEOF:5}
	clie := &fakeCli{stream:stream}
	ui := cli.NewMockUi()
	config := &commandhelper.ClientConfig{Client: clie, Theme: commandhelper.Default(true)}
	cmd2 := &cmd{UI: ui, config: config, OcyHelper: &commandhelper.OcyHelper{}}
	cmd2.init()
	code := cmd2.Run([]string{"-acct-repo=1/2", "-hash=1", "-branch=branch"})
	if code != 0 {
		t.Error("should return 0")
		t.Log(string(ui.ErrorWriter.Bytes()))
		t.Log(string(ui.OutputWriter.Bytes()))
	}
	var output []string
	for i := 0; i < 5; i++ {
		output = append(output, fmt.Sprintf("line number %d of %d", i, 5))
	}
	live := strings.Split(string(ui.OutputWriter.Bytes()), "\n")
	if diff := deep.Equal(output, live[:5]); diff != nil {
		t.Error(diff)
	}

	clie2 := &fakeCli{returnErr:true}
	ui2 := cli.NewMockUi()
	config = &commandhelper.ClientConfig{Client: clie2, Theme: commandhelper.Default(true)}
	cmd2 = &cmd{UI: ui2, config: config, OcyHelper: &commandhelper.OcyHelper{}}
	cmd2.init()
	code = cmd2.Run([]string{"-acct-repo=1/2", "-hash=1", "-branch=branch"})
	if code != 1 {
		t.Error("should fail as client returns an error")
	}
	out := string(ui2.ErrorWriter.Bytes())
	if out != "this is an error.\n" {
		t.Errorf("expected %s for output, got %s",  "this is an error.\n", out )
	}
	clie3 := &fakeCli{returnErr:false, stream:&fakeStream{fail:true}}
	ui3 := cli.NewMockUi()
	config = &commandhelper.ClientConfig{Client: clie3, Theme: commandhelper.Default(true)}
	cmd2 = &cmd{UI: ui3, config: config, OcyHelper: &commandhelper.OcyHelper{}}
	cmd2.init()
	code = cmd2.Run([]string{"-acct-repo=1/2", "-hash=1", "-branch=branch"})
	if code != 1 {
		t.Error("should fail as streaming client returns an error")
	}
	t.Log(string(ui3.ErrorWriter.Bytes()))
}

func TestCmd_Run_force(t *testing.T) {
	stream := &fakeStream{countBeforeEOF:5}
	clie := &fakeCli{stream:stream}
	ui := cli.NewMockUi()
	config := &commandhelper.ClientConfig{Client: clie, Theme: commandhelper.Default(true)}
	cmd2 := &cmd{UI: ui, config: config, OcyHelper: &commandhelper.OcyHelper{}}
	cmd2.init()
	code := cmd2.Run([]string{"-acct-repo=1/2", "-hash=1", "-branch=branch", "-force"})
	if code != 0 {
		t.Error("should return 0")
		t.Log(string(ui.ErrorWriter.Bytes()))
		t.Log(string(ui.OutputWriter.Bytes()))
	}
	expected := &pb.BuildReq{
		AcctRepo:"1/2",
		Hash:"1",
		Branch:"branch",
		Force:true,
	}
	if diff := deep.Equal(expected, clie.buildReq); diff != nil {
		t.Error(diff)
	}
}

func TestNew(t *testing.T) {
	cm := New(cli.NewMockUi())
	if cm.flags == nil {
		t.Error("init should have created a flags obj")
	}
}
package protobuf

import (
	"google.golang.org/grpc"
	"io"
)

//type BuildClient interface {
//	BuildInfo(ctx context.Context, in *Request, opts ...grpc.CallOption) (Build_BuildInfoClient, error)
//}

func NewFakeBuildClient(logLines []string) *fakeBuildClient {
	return &fakeBuildClient{
		outputLines: logLines,
	}
}

type fakeBuildClient struct {
	index int
	outputLines []string
	grpc.ClientStream
}

func (c *fakeBuildClient) CloseSend() error {
	return nil
}

func (c *fakeBuildClient) Recv() (*Response, error) {
	if c.index + 1 > len(c.outputLines) {
		return nil, io.EOF
	}
	resp := &Response{OutputLine: c.outputLines[c.index]}
	c.index++
	return resp, nil
}

package commandhelper

import (
	"github.com/level11consulting/ocelot/models/pb"
	"google.golang.org/grpc"
)

// CreateBuildClient dials the grpc server at the werker endpoints
func CreateBuildClient(m *pb.BuildRuntimeInfo) (pb.BuildClient, error) {
	//TODO: this is insecure
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(m.Ip+":"+m.GrpcPort, opts...)
	if err != nil {
		return nil, err
	}
	return pb.NewBuildClient(conn), nil
}

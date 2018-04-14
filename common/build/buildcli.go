package build


// CreateBuildClient dials the grpc server at the werker endpoints
func (m *BuildRuntimeInfo) CreateBuildClient(m *BuildRuntimeInfo) (protobuf.BuildClient, error) {
	//TODO: this is insecure
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err :=  grpc.Dial(m.Ip + ":" + m.GrpcPort, opts...)
	if err != nil {
		return nil, err
	}
	return protobuf.NewBuildClient(conn), nil
}


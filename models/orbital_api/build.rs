fn main() {
    tower_grpc_build::Config::new()
        .enable_server(true)
        .enable_client(true)
        .build(
            &[
                "../protos/build.proto",
                "../protos/credential.proto",
                "../protos/integration.proto",
                "../protos/organization.proto",
                "../protos/state.proto"
            ],
            &["../protos"],
        )
        .unwrap_or_else(|e| panic!("protobuf compilation failed: {}", e))
}
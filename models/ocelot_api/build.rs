fn main() {
    // Legacy api
    tower_grpc_build::Config::new()
        .enable_server(true)
        .enable_client(true)
        .build(
            &[
                "../../legacy/models/build.proto",
                "../../legacy/models/creds.proto",
                "../../legacy/models/guideocelot.proto",
                "../../legacy/models/storage.proto",
                "../../legacy/models/vcshandler.proto",
            ],
            &[
                "../../legacy/models",
                "../protos/vendor/grpc-gateway/third_party/googleapis",
            ],
        )
        .unwrap_or_else(|e| panic!("protobuf compilation failed: {}", e));
}

fn main() {
    // Legacy api
    tower_grpc_build::Config::new()
        .enable_server(true)
        .enable_client(true)
        .build(
            &[
                "../../../models/build.proto",
                "../../../models/creds.proto",
                "../../../models/guideocelot.proto",
                "../../../models/storage.proto",
                "../../../models/vcshandler.proto",
                // WONTFIX: Single namespace causes conflict w/ prost. We aren't implementing a werker anyway
                //"../../../models/werkerserver.proto",
            ],
            &[
                "../../../oldmodels",
                "../../../models/protos/vendor/grpc-gateway/third_party/googleapis",
            ],
        )
        .unwrap_or_else(|e| panic!("protobuf compilation failed: {}", e));
}

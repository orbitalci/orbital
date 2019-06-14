fn main() {
    // Legacy api
    tower_grpc_build::Config::new()
        .enable_server(true)
        .enable_client(true)
        .build(
            &[
                "../../../oldmodels/build.proto",
                "../../../oldmodels/creds.proto",
                "../../../oldmodels/guideocelot.proto",
                "../../../oldmodels/storage.proto",
                "../../../oldmodels/vcshandler.proto",
                // WONTFIX: Single namespace causes conflict w/ prost. We aren't implementing a werker anyway
                //"../../../models/werkerserver.proto",
            ],
            &["../../../oldmodels"],
        )
        .unwrap_or_else(|e| panic!("protobuf compilation failed: {}", e));
}

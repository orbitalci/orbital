fn main() {
    tower_grpc_build::Config::new()
        .enable_server(true)
        .enable_client(true)
        .build(
            &["../../oldmodels/build.proto"],
            &["../../oldmodels"],
        )
        .unwrap_or_else(|e| panic!("protobuf compilation failed: {}", e));

    tower_grpc_build::Config::new()
        .enable_server(true)
        .enable_client(true)
        .build(
            &["../../oldmodels/creds.proto"],
            &["../../oldmodels"],
        )
        .unwrap_or_else(|e| panic!("protobuf compilation failed: {}", e));

    // FIXME: Protos with services don't compile w/o modifications
    tower_grpc_build::Config::new()
        .enable_server(true)
        .enable_client(true)
        .build(
            &["../../oldmodels/guideocelot.proto"],
            &["../../oldmodels"],
        )
        .unwrap_or_else(|e| panic!("protobuf compilation failed: {}", e));

    tower_grpc_build::Config::new()
        .enable_server(true)
        .enable_client(true)
        .build(
            &["../../oldmodels/storage.proto"],
            &["../../oldmodels"],
        )
        .unwrap_or_else(|e| panic!("protobuf compilation failed: {}", e));

    tower_grpc_build::Config::new()
        .enable_server(true)
        .enable_client(true)
        .build(
            &["../../oldmodels/vcshandler.proto"],
            &["../../oldmodels"],
        )
        .unwrap_or_else(|e| panic!("protobuf compilation failed: {}", e));

    // FIXME: Protos with services don't compile w/o modifications
    tower_grpc_build::Config::new()
        .enable_server(true)
        .enable_client(true)
        .build(
            &["../../oldmodels/werkerserver.proto"],
            &["../../oldmodels"],
        )
        .unwrap_or_else(|e| panic!("protobuf compilation failed: {}", e));
}

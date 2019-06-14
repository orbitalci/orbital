fn main() {
    tower_grpc_build::Config::new()
        .enable_server(true)
        .enable_client(true)
        .build(
            &["../../models/build.proto"],
            &["../../models"],
        )
        .unwrap_or_else(|e| panic!("protobuf compilation failed: {}", e));

    tower_grpc_build::Config::new()
        .enable_server(true)
        .enable_client(true)
        .build(
            &["../../models/creds.proto"],
            &["../../models"],
        )
        .unwrap_or_else(|e| panic!("protobuf compilation failed: {}", e));

    // FIXME: Protos with services don't compile w/o modifications
    tower_grpc_build::Config::new()
        .enable_server(true)
        .enable_client(true)
        .build(
            &["../../models/guideocelot.proto"],
            &["../../models"],
        )
        .unwrap_or_else(|e| panic!("protobuf compilation failed: {}", e));

    tower_grpc_build::Config::new()
        .enable_server(true)
        .enable_client(true)
        .build(
            &["../../models/storage.proto"],
            &["../../models"],
        )
        .unwrap_or_else(|e| panic!("protobuf compilation failed: {}", e));

    tower_grpc_build::Config::new()
        .enable_server(true)
        .enable_client(true)
        .build(
            &["../../models/vcshandler.proto"],
            &["../../models"],
        )
        .unwrap_or_else(|e| panic!("protobuf compilation failed: {}", e));

    // FIXME: Protos with services don't compile w/o modifications
    tower_grpc_build::Config::new()
        .enable_server(true)
        .enable_client(true)
        .build(
            &["../../models/werkerserver.proto"],
            &["../../models"],
        )
        .unwrap_or_else(|e| panic!("protobuf compilation failed: {}", e));
}

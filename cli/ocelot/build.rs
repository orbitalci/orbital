fn main() {
    tower_grpc_build::Config::new()
        .enable_server(true)
        .enable_client(true)
        .build(
            &["../../legacy/models/build.proto"],
            &[
                "../../legacy/models",
                "../../models/protos/vendor/grpc-gateway/third_party/googleapis",
            ],
        )
        .unwrap_or_else(|e| panic!("protobuf compilation failed: {}", e));

    tower_grpc_build::Config::new()
        .enable_server(true)
        .enable_client(true)
        .build(
            &["../../legacy/models/creds.proto"],
            &[
                "../../legacy/models",
                "../../models/protos/vendor/grpc-gateway/third_party/googleapis",
            ],
        )
        .unwrap_or_else(|e| panic!("protobuf compilation failed: {}", e));

    tower_grpc_build::Config::new()
        .enable_server(true)
        .enable_client(true)
        .build(
            &["../../legacy/models/guideocelot.proto"],
            &[
                "../../legacy/models",
                "../../models/protos/vendor/grpc-gateway/third_party/googleapis",
            ],
        )
        .unwrap_or_else(|e| panic!("protobuf compilation failed: {}", e));

    tower_grpc_build::Config::new()
        .enable_server(true)
        .enable_client(true)
        .build(
            &["../../legacy/models/storage.proto"],
            &[
                "../../legacy/models",
                "../../models/protos/vendor/grpc-gateway/third_party/googleapis",
            ],
        )
        .unwrap_or_else(|e| panic!("protobuf compilation failed: {}", e));

    tower_grpc_build::Config::new()
        .enable_server(true)
        .enable_client(true)
        .build(
            &["../../legacy/models/vcshandler.proto"],
            &[
                "../../legacy/models",
                "../../models/protos/vendor/grpc-gateway/third_party/googleapis",
            ],
        )
        .unwrap_or_else(|e| panic!("protobuf compilation failed: {}", e));

    tower_grpc_build::Config::new()
        .enable_server(true)
        .enable_client(true)
        .build(
            &["../../legacy/models/werkerserver.proto"],
            &[
                "../../legacy/models",
                "../../models/protos/vendor/grpc-gateway/third_party/googleapis",
            ],
        )
        .unwrap_or_else(|e| panic!("protobuf compilation failed: {}", e));
}
fn main() -> Result<(), Box<dyn std::error::Error>> {
    tonic_build::compile_protos("../protos/build_metadata.proto")?;
    tonic_build::compile_protos("../protos/integration.proto")?;
    tonic_build::compile_protos("../protos/state.proto")?;

    /// This is a workaround for issue https://github.com/level11consulting/orbitalci/issues/229
    tonic_build::configure()
        .build_server(true)
        .build_client(true)
        .compile(
            &[
                "../protos/credential.proto",
                "../protos/credential_service.proto",
            ],
            &["../protos"],
        )?;

    tonic_build::configure()
        .build_server(true)
        .build_client(true)
        .compile(
            &["../protos/organization.proto"],
            &["../protos", "../protos/vendor/protobuf/src/google/protobuf"],
        )?;

    Ok(())
}

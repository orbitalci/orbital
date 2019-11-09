fn main() -> Result<(), Box<dyn std::error::Error>> {
    tonic_build::compile_protos("../protos/build_metadata.proto")?;
    tonic_build::compile_protos("../protos/integration.proto")?;
    tonic_build::compile_protos("../protos/state.proto")?;
    tonic_build::compile_protos("../protos/organization.proto")?;
    tonic_build::compile_protos("../protos/credential.proto")?;
    // This is a workaround for issue https://github.com/level11consulting/orbitalci/issues/229
    tonic_build::compile_protos("../protos/credential_service.proto")?;
    
    Ok(())
}

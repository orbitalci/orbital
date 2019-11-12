fn main() -> Result<(), Box<dyn std::error::Error>> {
    tonic_build::compile_protos("../protos/build_metadata.proto")?;
    tonic_build::compile_protos("../protos/integration.proto")?;
    tonic_build::compile_protos("../protos/state.proto")?;
    tonic_build::compile_protos("../protos/organization.proto")?;
    tonic_build::compile_protos("../protos/credential.proto")?;

    Ok(())
}

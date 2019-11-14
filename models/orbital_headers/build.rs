fn main() -> Result<(), Box<dyn std::error::Error>> {
    tonic_build::compile_protos("../protos/orbital_types.proto")?;
    tonic_build::compile_protos("../protos/build_meta.proto")?;
    tonic_build::compile_protos("../protos/notify.proto")?;
    tonic_build::compile_protos("../protos/organization.proto")?;
    tonic_build::compile_protos("../protos/secret.proto")?;
    tonic_build::compile_protos("../protos/code.proto")?;

    Ok(())
}

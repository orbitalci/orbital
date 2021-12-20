use chrono::Utc;
use git_meta::GitRepo;
use std::env;
use std::path::Path;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    tonic_build::compile_protos("protos/orbital_types.proto")?;
    tonic_build::compile_protos("protos/build_meta.proto")?;
    tonic_build::compile_protos("protos/notify.proto")?;
    tonic_build::compile_protos("protos/organization.proto")?;
    tonic_build::compile_protos("protos/secret.proto")?;
    tonic_build::compile_protos("protos/code.proto")?;

    println!("cargo:rerun-if-changed=protos/orbital_types.proto");
    println!("cargo:rerun-if-changed=protos/build_meta.proto");
    println!("cargo:rerun-if-changed=protos/notify.proto");
    println!("cargo:rerun-if-changed=protos/organization.proto");
    println!("cargo:rerun-if-changed=protos/secret.proto");
    println!("cargo:rerun-if-changed=protos/code.proto");

    let package_version = env::var("CARGO_PKG_VERSION")?;

    let cargo_build_root = env::var("CARGO_MANIFEST_DIR")?;

    // This is some manual walking to the root of the repo. Ew.
    let git_repo_dir = Path::new(&cargo_build_root);

    // This is all to get the commit id
    let git_repo =
        GitRepo::open(git_repo_dir.to_path_buf(), None, None).expect("Unable to open git repo");

    //// FIXME: The `cargo publish --dry-run` fails with this
    //let git_repo = GitRepo::open(git_repo_dir.to_path_buf(), None, None).unwrap_or_else(|_| {
    //    let cargo_publish_root = format!("{}/../../..", &cargo_build_root);
    //    let git_repo_cargo_publish_root = Path::new(&cargo_publish_root);
    //    GitRepo::open(git_repo_cargo_publish_root.to_path_buf(), None, None).unwrap()
    //});

    // Get git version
    let git_commit = git_repo.head.expect("No GitCommitMeta found").id[..12].to_string();

    // Get build datetime
    let now = Utc::now();

    let version_string = format!("{} ({}) {}", package_version, git_commit, now.to_rfc3339());

    println!("cargo:rustc-env=BUILD_VERSION={}", version_string);

    println!("cargo:rerun-if-changed=build.rs");
    Ok(())
}

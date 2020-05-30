use chrono::Utc;
use git_meta::git_info;
use std::env;
use std::path::Path;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let package_version = env::var("CARGO_PKG_VERSION")?;

    let cargo_build_root = env::var("CARGO_MANIFEST_DIR")?;

    // This is some manual walking to the root of the repo. Ew.
    let git_repo_dir = Path::new(&cargo_build_root)
        .parent()
        .unwrap()
        .parent()
        .unwrap()
        .parent()
        .unwrap()
        .parent()
        .unwrap();

    // Get git version
    let git_commit =
        git_info::get_git_info_from_path(&git_repo_dir, &None, &None)?.commit_id[..12].to_string();

    // Get build datetime
    let now = Utc::now();

    let version_string = format!("{} ({}) {}", package_version, git_commit, now.to_rfc3339());

    println!("cargo:rustc-env=BUILD_VERSION={}", version_string);

    Ok(())
}

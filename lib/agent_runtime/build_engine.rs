use git_meta;
use mktemp;
use std::error::Error;

/// Create a temporary directory on the host, and clone a repo
pub fn clone_repo(
    uri: &str,
    credentials: git_meta::GitCredentials,
) -> Result<mktemp::Temp, Box<dyn Error>> {
    git_meta::clone::clone_temp_dir(uri, credentials)
}

/// Helper functions for parsing local git repos and deriving Orbital accounting info
pub mod git_info;

/// This is the git reference that will be used for build requests
#[derive(Debug, Default)]
pub struct GitCommitContext {
    pub provider: String,
    pub branch: String,
    pub id: String,
    pub account: String,
    pub repo: String,
}

/// Parsed from a remote git uri
#[derive(Debug, PartialEq)]
pub struct GitSshRemote {
    user: String,
    provider: String,
    account: String,
    repo: String,
}

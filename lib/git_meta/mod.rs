/// Helper functions for cloning repos
pub mod clone;
/// Helper functions for parsing local git repos and deriving Orbital accounting info
pub mod git_info;

use git_url_parse::GitUrl;
use std::path::Path;

/// This is the git commit that will be used for build requests
#[derive(Debug, Default, Clone)]
pub struct GitCommitContext {
    pub branch: String,
    pub commit_id: String,
    pub git_url: GitUrl,
}

/// Types of supported git authentication
#[derive(Clone, Debug)]
pub enum GitCredentials<'a> {
    /// Public repo
    Public,
    /// Username, PrivateKey, PublicKey, Passphrase
    SshKey {
        username: String,
        public_key: Option<&'a Path>,
        private_key: &'a Path,
        passphrase: Option<&'a str>,
    },
    /// Username, Password
    UserPassPlaintext { username: String, password: String },
}

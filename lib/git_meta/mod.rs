/// Helper functions for cloning repos
pub mod clone;
/// Helper functions for parsing local git repos and deriving Orbital accounting info
pub mod git_info;

use git_url_parse::GitUrl;

/// This is the git commit that will be used for build requests
#[derive(Debug, Default, Clone)]
pub struct GitCommitContext {
    pub branch: String,
    pub commit_id: String,
    pub message: String,
    pub git_url: GitUrl,
}

/// Types of supported git authentication
#[derive(Clone, Debug)]
pub enum GitCredentials {
    /// Public repo
    Public,
    /// Username, PrivateKey, PublicKey, Passphrase
    SshKey {
        username: String,
        public_key: Option<String>,
        private_key: String,
        passphrase: Option<String>,
    },
    /// Username, Password
    BasicAuth { username: String, password: String },
}

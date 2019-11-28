/// Helper functions for cloning repos
pub mod clone;
/// Helper functions for parsing local git repos and deriving Orbital accounting info
pub mod git_info;

use std::path::Path;

/// This is the git reference that will be used for build requests
#[derive(Debug, Default)]
pub struct GitCommitContext {
    pub provider: String,
    pub branch: String,
    pub id: String,
    pub account: String,
    pub repo: String,
    pub uri: String,
}

/// Parsed from a remote git uri
#[derive(Debug, PartialEq)]
pub struct GitSshRemote {
    user: String,
    provider: String,
    account: String,
    repo: String,
}

/// Types of supported git authentication
#[derive(Clone, Debug)]
pub enum GitCredentials {
    /// Public repo
    Public,
    /// Username, PrivateKey, PublicKey, Passphrase
    SshKey {
        username: String,
        public_key: Option<&'static Path>,
        private_key: &'static Path,
        passphrase: Option<&'static str>,
    },
    /// Username, Password
    UserPassPlaintext {
        username: String,
        password: String,
    },
}

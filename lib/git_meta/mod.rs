/// Helper functions for cloning repos
pub mod clone;
/// Helper functions for parsing local git repos and deriving Orbital accounting info
pub mod git_info;

use git_url_parse::GitUrl;

use git2::{Cred, RemoteCallbacks};
use mktemp::Temp;
use std::fs::File;
use std::io::prelude::*;

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

fn build_remote_callback<'a>(credentials: GitCredentials) -> RemoteCallbacks<'a> {
    match credentials {
        GitCredentials::Public => RemoteCallbacks::new(),

        GitCredentials::SshKey {
            username,
            public_key,
            private_key,
            passphrase,
        } => {
            let mut callbacks = RemoteCallbacks::new();
            // Write private key to temp file

            let privkey_file =
                Temp::new_file().expect("Unable to create temp file for private key");

            let mut privkey_fd = File::create(privkey_file.as_path()).unwrap();
            let _ = privkey_fd.write_all(private_key.as_bytes());

            &callbacks.credentials(move |_, _, _| {
                // Do some Option re-wrapping stuff bc lifetimes
                match (public_key.clone(), passphrase.clone()) {
                    (None, None) => {
                        Ok(Cred::ssh_key(&username, None, privkey_file.as_path(), None)
                            .expect("Could not create credentials object for ssh key"))
                    }
                    (None, Some(pp)) => Ok(Cred::ssh_key(
                        &username,
                        None,
                        privkey_file.as_path(),
                        Some(pp.as_ref()),
                    )
                    .expect("Could not create credentials object for ssh key")),
                    (Some(pk), None) => {
                        // Write public key to temp file
                        let pubkey_file =
                            Temp::new_file().expect("Unable to create temp file for public key");
                        let mut pubkey_fd = File::create(pubkey_file.as_path()).unwrap();
                        let _ = pubkey_fd.write_all(pk.as_bytes());

                        Ok(Cred::ssh_key(
                            &username,
                            Some(pubkey_file.as_path()),
                            privkey_file.as_path(),
                            None,
                        )
                        .expect("Could not create credentials object for ssh key"))
                    }
                    (Some(pk), Some(pp)) => {
                        // Write public key to temp file
                        let pubkey_file =
                            Temp::new_file().expect("Unable to create temp file for public key");
                        let mut pubkey_fd = File::create(pubkey_file.as_path()).unwrap();
                        let _ = pubkey_fd.write_all(pk.as_bytes());

                        Ok(Cred::ssh_key(
                            &username,
                            Some(pubkey_file.as_path()),
                            privkey_file.as_path(),
                            Some(pp.as_ref()),
                        )
                        .expect("Could not create credentials object for ssh key"))
                    }
                }
            });

            callbacks
        }

        GitCredentials::BasicAuth { username, password } => RemoteCallbacks::new(),
    }
}

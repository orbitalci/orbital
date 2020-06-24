use crate::GitCredentials;
use anyhow::Result;
use git2::{build::RepoBuilder, Cred, FetchOptions, RemoteCallbacks};
use log::debug;
use mktemp::Temp;
use std::fs::File;
use std::io::prelude::*;

// TODO: Need a way to switch between a public and private repo
// Idea: Create an enum:

/// Create a temporary directory with mktemp, clone given uri into it.
/// Return mktemp directory, which will delete when out of scope
pub fn clone_temp_dir(uri: &str, branch: &str, credentials: GitCredentials) -> Result<Temp> {
    let temp_dir = Temp::new_dir()?;

    debug!("Temp dir path: {:?}", &temp_dir.as_path());
    debug!("GitCredentials: {:?}", &credentials);

    match credentials {
        GitCredentials::Public => {
            debug!("Cloning a public repo");

            let mut builder = RepoBuilder::new();
            let callbacks = RemoteCallbacks::new();
            let mut fetch_options = FetchOptions::new();

            fetch_options.remote_callbacks(callbacks);
            builder.fetch_options(fetch_options);
            builder.branch(branch);

            let _repo = match builder.clone(uri, &temp_dir.as_path()) {
                Ok(repo) => repo,
                Err(e) => panic!("failed to clone: {}", e),
            };
        }
        GitCredentials::SshKey {
            username,
            public_key,
            private_key,
            passphrase,
        } => {
            debug!("Cloning a private repo with ssh keys");

            let mut builder = RepoBuilder::new();
            let mut callbacks = RemoteCallbacks::new();
            let mut fetch_options = FetchOptions::new();

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

            fetch_options.remote_callbacks(callbacks);
            builder.fetch_options(fetch_options);
            builder.branch(branch);

            let _repo = match builder.clone(uri, &temp_dir.as_path()) {
                Ok(repo) => repo,
                Err(e) => panic!("failed to clone: {}", e),
            };
        }

        GitCredentials::BasicAuth { username, password } => {
            debug!("Cloning a private repo with basic auth");

            let mut builder = RepoBuilder::new();
            let mut callbacks = RemoteCallbacks::new();
            let mut fetch_options = FetchOptions::new();

            &callbacks.credentials(|_, _, _| {
                let userpass = Cred::userpass_plaintext(&username, &password)
                    .expect("Could not create credentials object for userpass_plaintext");
                Ok(userpass)
            });

            fetch_options.remote_callbacks(callbacks);
            builder.fetch_options(fetch_options);
            builder.branch(branch);

            let _repo = match builder.clone(uri, &temp_dir.as_path()) {
                Ok(repo) => repo,
                Err(e) => panic!("failed to clone: {}", e),
            };
        }
    }

    Ok(temp_dir)
}

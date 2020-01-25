use crate::GitCredentials;
use anyhow::Result;
use git2::{build::RepoBuilder, Cred, FetchOptions, RemoteCallbacks};
use log::debug;
use mktemp::Temp;

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

            //let _repo = match Repository::clone(uri, &temp_dir.as_path()) {
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

            &callbacks.credentials(|_, _, _| {
                Ok(
                    Cred::ssh_key(&username, public_key, private_key, passphrase)
                        .expect("Could not create credentials object for ssh key"),
                )
            });

            fetch_options.remote_callbacks(callbacks);
            builder.fetch_options(fetch_options);
            builder.branch(branch);

            let _repo = match builder.clone(uri, &temp_dir.as_path()) {
                Ok(repo) => repo,
                Err(e) => panic!("failed to clone: {}", e),
            };
        }

        GitCredentials::UserPassPlaintext { username, password } => {
            debug!("Cloning a private repo with userpass");

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

            //let _repo = match Repository::clone(uri, &temp_dir.as_path()) {
            let _repo = match builder.clone(uri, &temp_dir.as_path()) {
                Ok(repo) => repo,
                Err(e) => panic!("failed to clone: {}", e),
            };
        }
    }

    Ok(temp_dir)
}

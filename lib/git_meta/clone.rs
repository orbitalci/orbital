use crate::GitCredentials;
use anyhow::Result;
use git2::{build::RepoBuilder, Cred, FetchOptions, RemoteCallbacks};
use log::debug;
use mktemp::Temp;
use std::fs::File;
use std::io::prelude::*;
use std::path::{Path, PathBuf};
use std::process::{Command, Stdio};

// TODO: Need a way to switch between a public and private repo
// Idea: Create an enum:

/// Create a temporary directory with mktemp, clone given uri into it.
/// Return mktemp directory, which will delete when out of scope
pub fn clone_temp_dir(
    uri: &str,
    branch: Option<&str>,
    credentials: GitCredentials,
    target_dir: &Path,
) -> Result<()> {
    debug!("Temp dir path: {:?}", &target_dir);
    debug!("GitCredentials: {:?}", &credentials);

    let git_callbacks = super::build_remote_callback(credentials.clone());

    match credentials {
        GitCredentials::Public => {
            debug!("Cloning a public repo");

            let mut builder = RepoBuilder::new();
            let callbacks = git_callbacks;
            let mut fetch_options = FetchOptions::new();

            fetch_options.remote_callbacks(callbacks);
            builder.fetch_options(fetch_options);

            if let Some(b) = branch {
                builder.branch(b);
            }

            let _repo = match builder.clone(uri, &target_dir) {
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
            let callbacks = git_callbacks;
            let mut fetch_options = FetchOptions::new();

            fetch_options.remote_callbacks(callbacks);
            builder.fetch_options(fetch_options);

            if let Some(b) = branch {
                builder.branch(b);
            }

            let _repo = match builder.clone(uri, &target_dir) {
                Ok(repo) => repo,
                Err(e) => panic!("failed to clone: {}", e),
            };
        }

        GitCredentials::BasicAuth { username, password } => {
            debug!("Cloning a private repo with basic auth");

            let mut builder = RepoBuilder::new();
            let callbacks = git_callbacks;
            let mut fetch_options = FetchOptions::new();

            fetch_options.remote_callbacks(callbacks);
            builder.fetch_options(fetch_options);

            if let Some(b) = branch {
                builder.branch(b);
            }

            let _repo = match builder.clone(uri, &target_dir) {
                Ok(repo) => repo,
                Err(e) => panic!("failed to clone: {}", e),
            };
        }
    }

    Ok(())
}

// Shallow clone
/// This requires `git` and `ssh` to be installed
pub fn shell_shallow_clone(
    uri: &str,
    branch: Option<&str>,
    credentials: GitCredentials,
    target_dir: &Path,
) -> Result<()> {
    // Let's not assume that the URI has been parsed, ok?
    let parsed_uri = git_url_parse::GitUrl::parse(uri)?.trim_auth();

    match credentials {
        GitCredentials::BasicAuth { username, password } => {
            let mut cli_remote_url = parsed_uri.clone();
            cli_remote_url.user = Some(username);
            cli_remote_url.token = Some(password);

            let shell_clone_command = Command::new("git")
                .arg("clone")
                .arg(format!("{}", cli_remote_url))
                .arg(format!("{}", target_dir.to_str().unwrap()))
                .arg("--depth=1")
                .stdout(Stdio::piped())
                .stderr(Stdio::null())
                .spawn()
                .expect("Failed to run git clone");

            let _clone_out = shell_clone_command
                .wait_with_output()
                .expect("Failed to open stdout");
        }
        GitCredentials::Public => {
            let shell_clone_command = Command::new("git")
                .arg("clone")
                .arg(format!("{}", parsed_uri))
                .arg(format!("{}", target_dir.to_str().unwrap()))
                .arg("--depth=1")
                .stdout(Stdio::piped())
                .stderr(Stdio::null())
                .spawn()
                .expect("Failed to run git clone");

            let _clone_out = shell_clone_command
                .wait_with_output()
                .expect("Failed to open stdout");
        }
        GitCredentials::SshKey {
            username,
            public_key,
            private_key,
            passphrase,
        } => {
            // write private key to a temp file
            let privkey_file =
                Temp::new_file().expect("unable to create temp file for private key");

            let mut privkey_fd = File::create(privkey_file.as_path()).unwrap();
            let _ = privkey_fd.write_all(private_key.as_bytes());

            let shell_clone_command = Command::new("git")
                .arg("clone")
                .arg(format!("{}", parsed_uri))
                .arg(format!("{}", target_dir.to_str().unwrap()))
                .arg("--depth=1")
                .arg("--config")
                .arg(format!(
                    "core.sshcommand=\"ssh -i {privkey_path}\"",
                    privkey_path = privkey_file.as_path().display()
                ))
                .stdout(Stdio::piped())
                .stderr(Stdio::null())
                .spawn()
                .expect("failed to run git clone");

            let _clone_out = shell_clone_command
                .wait_with_output()
                .expect("Failed to open stdout");
        }
    }

    Ok(())
}

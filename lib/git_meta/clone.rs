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

    match credentials {
        GitCredentials::Public => {
            debug!("Cloning a public repo");

            let mut builder = RepoBuilder::new();
            let callbacks = RemoteCallbacks::new();
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
            let mut callbacks = RemoteCallbacks::new();
            let mut fetch_options = FetchOptions::new();

            &callbacks.credentials(|_, _, _| {
                let userpass = Cred::userpass_plaintext(&username, &password)
                    .expect("Could not create credentials object for userpass_plaintext");
                Ok(userpass)
            });

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

            let clone_out = shell_clone_command.stdout.expect("Failed to open stdout");
            //println!("{:?}", clone_out)
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

            let clone_out = shell_clone_command.stdout.expect("Failed to open stdout");
            //println!("{:?}", clone_out)
        }
        GitCredentials::SshKey {
            username,
            public_key,
            private_key,
            passphrase,
        } => {
            // write private key to a temp file
            let privkey_file =
                Temp::new_file().expect("Unable to create temp file for private key");

            let mut privkey_fd = File::create(privkey_file.as_path()).unwrap();
            let _ = privkey_fd.write_all(private_key.as_bytes());

            //let shell_clone_command = format!(
            //    "git clone {url} {dir} --depth=1 --config core.sshCommand=\"ssh -i {privkey_path}\"",
            //    url=cli_remote_url,
            //    dir=target_dir.to_str().unwrap(),
            //    privkey_path=privkey_file.as_path().display(),
            //);

            let shell_clone_command = Command::new("git")
                .arg("clone")
                .arg(format!("{}", parsed_uri))
                .arg(format!("{}", target_dir.to_str().unwrap()))
                .arg("--depth=1")
                .arg("--config")
                .arg(format!(
                    "core.sshCommand=\"ssh -i {privkey_path}\"",
                    privkey_path = privkey_file.as_path().display()
                ))
                .stdout(Stdio::piped())
                .stderr(Stdio::null())
                .spawn()
                .expect("Failed to run git clone");

            let clone_out = shell_clone_command.stdout.expect("Failed to open stdout");
            //println!("{:?}", clone_out)
        }
    }
    // git clone git@github.com:tjtelan/orbitalci.git --depth 1 --config core.sshCommand="ssh -i ~/.ssh/id_ed25519"

    //let mut git_config = PathBuf::new();
    //git_config.push(target_dir);
    //git_config.push(".git");
    //git_config.push("config");

    //let _file = File::create(git_config)?;

    Ok(())
}

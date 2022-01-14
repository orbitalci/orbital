pub mod docker;

use log::debug;
use std::env;
use std::path::PathBuf;

/// Default volume mount mapping for host Docker into container for Docker-in-Docker builds
pub const DOCKER_SOCKET_VOLMAP: &str = "/var/run/docker.sock:/var/run/docker.sock";
/// Default working directory for staging repo code inside container
pub const ORBITAL_CONTAINER_WORKDIR: &str = "/orbital-work";

// Below is copied from orbital_cli_subcommand crate

/// Wrapper function for `kv_csv_parser` to specifically handle env vars for `shiplift`
pub fn parse_envs_input(user_input: &Option<String>) -> Option<Vec<&str>> {
    let envs = kv_csv_parser(user_input);
    debug!("Env vars to set: {:?}", envs);
    envs
}

/// Returns a `Path` of the current working directory.
pub fn get_current_workdir() -> PathBuf {
    let pathbuf = match env::current_dir() {
        Ok(p) => p,
        Err(_) => panic!("Could not get current working directory"),
    };

    debug!("Current workdir on host: {:?}", &pathbuf);
    pathbuf
}

/// Wrapper function for `kv_csv_parser` to specifically handle volume mounts for `shiplift`
/// Automatically add in the docker socket as defined by `orbital_agent::DOCKER_SOCKET_VOLMAP`. If we don't pass in any other volumes
///
/// For now, also assume passing in the current working directory as well
pub fn parse_volumes_input(user_input: &Option<String>) -> Option<Vec<&str>> {
    let vols = match kv_csv_parser(user_input) {
        Some(v) => {
            let mut new_vec: Vec<&str> = vec![DOCKER_SOCKET_VOLMAP];
            new_vec.extend(v.clone());
            Some(new_vec)
        }
        None => {
            let mut new_vec: Vec<&str> = Vec::new();
            new_vec.push(DOCKER_SOCKET_VOLMAP);

            // There's got to be a better way to handle this...
            // https://stackoverflow.com/a/30527289/1672638
            new_vec.push(Box::leak(
                format!(
                    "{}:{}",
                    &get_current_workdir().display(),
                    ORBITAL_CONTAINER_WORKDIR,
                )
                .into_boxed_str(),
            ));
            Some(new_vec)
        }
    };
    debug!("Volumes to mount: {:?}", &vols);
    vols
}

/// Returns an `Option<Vec<&str>>` after parsing a comma-separated string from the cli
pub fn kv_csv_parser(kv_str: &Option<String>) -> Option<Vec<&str>> {
    debug!("Parsing Option<String> input: {:?}", &kv_str);
    match kv_str {
        Some(n) => {
            let kv_vec: Vec<&str> = n.split(',').collect();
            Some(kv_vec)
        }
        None => None,
    }
}

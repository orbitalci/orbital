/// Agent api for building
pub mod build_engine;

/// Docker engine api wrapper
pub mod docker;
/// Vault api wrapper
pub mod vault;
/// Default volume mount mapping for host Docker into container for Docker-in-Docker builds
pub const DOCKER_SOCKET_VOLMAP: &str = "/var/run/docker.sock:/var/run/docker.sock";
/// Default working directory for staging repo code inside container
pub const ORBITAL_CONTAINER_WORKDIR: &str = "/orbital-work";

use log::debug;
use std::error::Error;
use std::path::PathBuf;
use std::{env, fmt};
use thiserror::Error;

// Leaving this here for when we can focus on non-docker workflows
//enum AgentRuntimeType {
//    Host,
//    Docker,
//}

#[derive(Debug, Error)]
pub struct AgentRuntimeError {
    details: String,
}

impl AgentRuntimeError {
    pub fn new(msg: &str) -> AgentRuntimeError {
        AgentRuntimeError {
            details: msg.to_string(),
        }
    }
}

impl fmt::Display for AgentRuntimeError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}", self.details)
    }
}

impl From<Box<dyn Error>> for AgentRuntimeError {
    fn from(error: Box<dyn Error>) -> Self {
        AgentRuntimeError::new(&error.to_string())
    }
}

/// Generate a tracable string for Docker containers
pub fn generate_unique_build_id(org: &str, repo: &str, commit: &str, id: &str) -> String {
    // Arbitrary max lengths
    let org_print = if org.len() > 20 { &org[0..19] } else { org };
    let repo_print = if repo.len() > 20 { &repo[0..19] } else { repo };
    let commit_print = &commit[0..6];

    format!(
        "{org}_{repo}_{commit}_{id}",
        org = org_print,
        repo = repo_print,
        commit = commit_print,
        id = id
    )
}

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
/// Automatically add in the docker socket as defined by `agent_runtime::DOCKER_SOCKET_VOLMAP`. If we don't pass in any other volumes
///
/// For now, also assume passing in the current working directory as well
pub fn parse_volumes_input(user_input: &Option<String>) -> Option<Vec<&str>> {
    let vols = match kv_csv_parser(user_input) {
        Some(v) => {
            let mut new_vec: Vec<&str> = Vec::new();
            new_vec.push(crate::DOCKER_SOCKET_VOLMAP);
            new_vec.extend(v.clone());
            Some(new_vec)
        }
        None => {
            let mut new_vec: Vec<&str> = Vec::new();
            new_vec.push(crate::DOCKER_SOCKET_VOLMAP);

            // There's got to be a better way to handle this...
            // https://stackoverflow.com/a/30527289/1672638
            new_vec.push(Box::leak(
                format!(
                    "{}:{}",
                    &get_current_workdir().display(),
                    crate::ORBITAL_CONTAINER_WORKDIR,
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
            let kv_vec: Vec<&str> = n.split(",").collect();
            return Some(kv_vec);
        }
        None => return None,
    }
}

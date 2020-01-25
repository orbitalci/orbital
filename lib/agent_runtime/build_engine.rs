use crate::docker;
use crate::AgentRuntimeError;
use anyhow::Result;
use config_parser;
use git_meta;
use log::debug;
use mktemp;
use std::path::Path;
use std::time::Duration;

/// Create a temporary directory on the host, and clone a repo
pub fn clone_repo(
    uri: &str,
    branch: &str,
    credentials: git_meta::GitCredentials,
) -> Result<mktemp::Temp> {
    git_meta::clone::clone_temp_dir(uri, branch, credentials)
}

/// Load orb.yml from a filepath
pub fn load_orb_config(path: &Path) -> Result<config_parser::OrbitalConfig> {
    config_parser::yaml::load_orb_yaml(path)
}

/// Pull a docker image using the host docker engine
pub fn docker_container_pull(image: &str) -> Result<()> {
    match docker::container_pull(image) {
        Ok(ok) => Ok(ok), // The successful result doesn't matter
        Err(_) => Err(AgentRuntimeError::new(&format!("Could not pull image {}", image)).into()),
    }
}

/// Create a docker container
pub fn docker_container_create(
    image: &str,
    envs: Option<Vec<&str>>,
    volumes: Option<Vec<&str>>,
    timeout: Duration,
) -> Result<String> {
    let timeout_as_seconds = format!("{}s", timeout.as_secs());
    let default_command_w_timeout = vec!["sleep", &timeout_as_seconds];
    match docker::container_create(image, default_command_w_timeout, envs, volumes) {
        Ok(container_id) => Ok(container_id),
        Err(_) => Err(AgentRuntimeError::new(&format!("Could not create image {}", &image)).into()),
    }
}

/// Start a docker container
pub fn docker_container_start(container_id: &str) -> Result<()> {
    match docker::container_start(container_id) {
        Ok(ok) => Ok(ok), // The successful result doesn't matter
        Err(_) => Err(AgentRuntimeError::new(&format!(
            "Could not start container_id {}",
            container_id
        ))
        .into()),
    }
}

/// Stop a docker container
pub fn docker_container_stop(container_id: &str) -> Result<()> {
    match docker::container_stop(container_id) {
        Ok(ok) => Ok(ok), // The successful result doesn't matter
        Err(_) => Err(AgentRuntimeError::new(&format!(
            "Could not start container_id {}",
            container_id
        ))
        .into()),
    }
}

/// Loop over commands, exec into docker container
pub fn docker_container_exec(container_id: &str, commands: Vec<String>) -> Result<()> {
    for command in commands.iter() {
        // Build the exec string
        let wrapped_command = format!("{} | tee -a /proc/1/fd/1", &command);

        let container_command = vec!["/bin/sh", "-c", wrapped_command.as_ref()];

        match docker::container_exec(container_id, container_command.clone()) {
            Ok(output) => {
                debug!("Command: {:?}", &command);
                debug!("Output: {:?}", &output);
                output
            }
            Err(_) => {
                return Err(AgentRuntimeError::new(&format!(
                    "Could not exec into container {}",
                    &container_id
                ))
                .into())
            }
        };
    }

    Ok(())
}

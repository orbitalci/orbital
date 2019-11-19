use crate::docker;
use crate::AgentRuntimeError;
use config_parser;
use git_meta;
use log::debug;
use mktemp;
use std::error::Error;
use std::path::Path;
use std::time::Duration;

/// Create a temporary directory on the host, and clone a repo
pub fn clone_repo(
    uri: &str,
    credentials: git_meta::GitCredentials,
) -> Result<mktemp::Temp, Box<dyn Error>> {
    git_meta::clone::clone_temp_dir(uri, credentials)
}

/// Load orb.yml from a filepath
pub fn load_orb_config(path: &Path) -> Result<config_parser::OrbitalConfig, Box<dyn Error>> {
    config_parser::yaml::load_orb_yaml(path)
}

/// Pull a docker image using the host docker engine
pub fn docker_container_pull(image: &str) -> Result<(), Box<dyn Error>> {
    match docker::container_pull(image) {
        Ok(ok) => Ok(ok), // The successful result doesn't matter
        Err(_) => Err(Box::new(AgentRuntimeError::new(&format!(
            "Could not pull image {}",
            image
        )))),
    }
}

/// Create a docker container
pub fn docker_container_create(
    image: &str,
    envs: Option<Vec<&str>>,
    volumes: Option<Vec<&str>>,
    timeout: Duration,
) -> Result<String, Box<dyn Error>> {
    let timeout_as_seconds = format!("{}s", timeout.as_secs());
    let default_command_w_timeout = vec!["sleep", &timeout_as_seconds];
    match docker::container_create(image, default_command_w_timeout, envs, volumes) {
        Ok(container_id) => Ok(container_id),
        Err(_) => Err(Box::new(AgentRuntimeError::new(&format!(
            "Could not create image {}",
            &image
        )))),
    }
}

/// Start a docker container
pub fn docker_container_start(image: &str) -> Result<(), Box<dyn Error>> {
    match docker::container_start(image) {
        Ok(ok) => Ok(ok), // The successful result doesn't matter
        Err(_) => Err(Box::new(AgentRuntimeError::new(&format!(
            "Could not pull image {}",
            image
        )))),
    }
}

/// Loop over commands, exec into docker container
pub fn docker_container_exec(
    container_id: &str,
    commands: Vec<String>,
) -> Result<(), Box<dyn Error>> {
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
                return Err(Box::new(AgentRuntimeError::new(&format!(
                    "Could not exec into container {}",
                    &container_id
                ))))
            }
        };
    }

    Ok(())
}

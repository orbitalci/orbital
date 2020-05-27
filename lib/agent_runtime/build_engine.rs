use crate::docker::{self, OrbitalContainerSpec};
use crate::AgentRuntimeError;
use anyhow::Result;
use config_parser;
use git_meta;
use log::{debug, info};
use mktemp;
use std::path::Path;

use serde_json::value::Value;
use tokio::sync::mpsc;

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

/// Load config from str
pub fn load_orb_config_from_str(config: &str) -> Result<config_parser::OrbitalConfig> {
    config_parser::yaml::load_orb_yaml_from_str(config)
}

/// Pull a docker image using the host docker engine
pub fn docker_container_pull(orb_build_spec: &OrbitalContainerSpec) -> Result<()> {
    match docker::container_pull(&orb_build_spec.image) {
        Ok(ok) => Ok(ok), // The successful result doesn't matter
        Err(_) => Err(AgentRuntimeError::new(&format!(
            "Could not pull image {}",
            &orb_build_spec.image
        ))
        .into()),
    }
}

pub async fn docker_container_pull_async(
    orb_build_spec: OrbitalContainerSpec<'_>,
) -> Result<mpsc::UnboundedReceiver<Value>> {
    docker::container_pull_async(orb_build_spec.image.to_string()).await
}

/// Create a docker container
pub fn docker_container_create(orb_build_spec: &OrbitalContainerSpec) -> Result<String> {
    let timeout_as_seconds = format!("{}s", orb_build_spec.timeout.unwrap().as_secs());
    let default_command_w_timeout = vec!["sleep", &timeout_as_seconds];

    let mut orb_build_spec_w_timeout = orb_build_spec.clone();
    orb_build_spec_w_timeout.command = default_command_w_timeout;

    match docker::container_create(orb_build_spec_w_timeout) {
        Ok(container_id) => {
            info!("Created container id: {:?}", container_id);
            Ok(container_id)
        }
        Err(_) => Err(AgentRuntimeError::new(&format!(
            "Could not create image {}",
            &orb_build_spec.image
        ))
        .into()),
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
            "Could not stop container_id {}",
            container_id
        ))
        .into()),
    }
}

// FIXME: Possibly change this to only run single commands. So timestamping can be handled outside
// TODO: This will also need to accept some channel to pass to docker::container_exec
/// Loop over commands, exec into docker container
pub fn docker_container_exec(container_id: &str, commands: Vec<String>) -> Result<String> {
    let mut exec_output: Vec<String> = Vec::new();

    for command in commands.iter() {
        // Build the exec string
        let wrapped_command = format!("{} | tee -a /proc/1/fd/1", &command);

        let container_command = vec!["/bin/sh", "-c", wrapped_command.as_ref()];

        // Print the command in the output we want to store in database
        &mut exec_output.push(format!("Command: {:?}\n", &command.clone()));
        match docker::container_exec(container_id, container_command.clone()) {
            Ok(output) => {
                debug!("Command: {:?}", &command);
                debug!("Output: {:?}", &output);
                &mut exec_output.extend(output.clone());
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

    Ok(exec_output.join(""))
}

pub async fn docker_container_exec_async(
    container_id: String,
    commands: Vec<String>,
) -> Result<mpsc::UnboundedReceiver<String>> {
    let (tx, rx) = mpsc::unbounded_channel();

    tokio::spawn(async move {
        for command in commands.iter() {
            // Build the exec string
            let wrapped_command = format!("{} | tee -a /proc/1/fd/1", command);

            let container_command = vec!["/bin/sh".to_string(), "-c".to_string(), wrapped_command];

            let mut exec_rx =
                docker::container_exec_async(container_id.clone(), container_command.clone())
                    .await
                    .unwrap();

            tx.send(format!("Command: {:?}\n", &command)).unwrap();

            while let Some(command_output) = exec_rx.recv().await {
                tx.send(command_output).unwrap();
            }
        }
    });

    Ok(rx)
}

use crate::orbital_utils::config_parser;
use crate::orbital_utils::exec_runtime::docker::{self, OrbitalContainerSpec};
use crate::orbital_utils::orbital_agent::AgentRuntimeError;
use color_eyre::eyre::Result;
use git_meta::GitRepo;
use tracing::info;
use std::path::Path;

use serde_json::value::Value;
use tokio::sync::mpsc;

/// TODO: Hang all of the bare functions off of Agent
//pub struct Agent;

/// Create a temporary directory on the host, and clone a repo
pub fn clone_repo<S: AsRef<str>>(
    uri: S,
    branch: Option<S>,
    credentials: Option<git_meta::GitCredentials>,
    target_dir: &Path,
) -> Result<()> {
    let git_repo = match (branch, credentials) {
        (Some(b), Some(c)) => GitRepo::new(uri)
            .expect("Cannot create GitRepo")
            .with_branch(Some(b.as_ref().to_string()))
            .with_credentials(Some(c)),
        (Some(b), None) => GitRepo::new(uri)
            .expect("Cannot create GitRepo")
            .with_branch(Some(b.as_ref().to_string())),
        (None, Some(c)) => GitRepo::new(uri)
            .expect("Cannot create GitRepo")
            .with_credentials(Some(c)),
        (None, None) => GitRepo::new(uri).expect("Cannot create GitRepo"),
    };

    git_repo
        .to_clone()
        .git_clone(target_dir)
        .expect("Failed to clone");

    Ok(())
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
pub async fn docker_container_pull(
    orb_build_spec: OrbitalContainerSpec<'_>,
) -> Result<mpsc::UnboundedReceiver<Value>> {
    docker::container_pull(orb_build_spec.image.to_string()).await
}

/// Create a docker container
pub async fn docker_container_create(orb_build_spec: &OrbitalContainerSpec<'_>) -> Result<String> {
    let timeout_as_seconds = format!("{}s", orb_build_spec.timeout.unwrap().as_secs());
    let default_command_w_timeout = vec!["sleep", &timeout_as_seconds];

    let mut orb_build_spec_w_timeout = orb_build_spec.clone();
    orb_build_spec_w_timeout.command = default_command_w_timeout;

    match docker::container_create(orb_build_spec_w_timeout).await {
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
pub async fn docker_container_start(container_id: &str) -> Result<()> {
    match docker::container_start(container_id).await {
        Ok(ok) => Ok(ok), // The successful result doesn't matter
        Err(_) => Err(AgentRuntimeError::new(&format!(
            "Could not start container_id {}",
            container_id
        ))
        .into()),
    }
}

/// Stop a docker container
pub async fn docker_container_stop(container_id: &str) -> Result<()> {
    match docker::container_stop(container_id).await {
        Ok(ok) => Ok(ok), // The successful result doesn't matter
        Err(_) => Err(AgentRuntimeError::new(&format!(
            "Could not stop container_id {}",
            container_id
        ))
        .into()),
    }
}

// FIXME: Possibly change this to only run single commands. So timestamping can be handled outside
/// Loop over commands, exec into docker container
pub async fn docker_container_exec(
    container_id: String,
    commands: Vec<String>,
) -> Result<mpsc::UnboundedReceiver<String>> {
    let (tx, rx) = mpsc::unbounded_channel();

    tokio::spawn(async move {
        for command in commands.iter() {
            // Build the exec string
            let wrapped_command = format!("{} | tee -a /proc/1/fd/1", command);

            let container_command = vec!["/bin/sh", "-c", &wrapped_command];

            let mut exec_rx = docker::container_exec(container_id.clone(), container_command)
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

pub async fn docker_container_logs(
    container_id: String,
) -> Result<mpsc::UnboundedReceiver<String>> {
    docker::container_logs(container_id).await
}

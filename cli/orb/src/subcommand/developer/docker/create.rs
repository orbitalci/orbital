use crate::{developer::docker::SubcommandOption, GlobalOption, SubcommandError};
use orbital_exec_runtime::{self, docker, docker::OrbitalContainerSpec};
//use log::debug;
use anyhow::Result;
use structopt::StructOpt;

use rand::distributions::Alphanumeric;
use rand::{thread_rng, Rng};
use std::time::Duration;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    /// Add env vars to build. Comma-separated with no spaces. ex. "key1=var1,key2=var2"
    #[structopt(long, short)]
    env: Option<String>,

    /// Add volume mapping from host to container. Comma-separated with no spaces. ex. "/host/path1:/container/path1,/host/path2:/container/path2"
    #[structopt(long, short)]
    volume: Option<String>,

    /// Docker image. If no tag provided, :latest will be assumed
    image: String,

    /// String command to execute in container. Will naively split on whitespace.
    command: String,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<()> {
    // FIXME
    // This is going to be a stupid parsed command on whitespace only.
    // Embedded commands with quotes, $(), or backtics not expected to work with this parsing
    //let command_vec_slice: Vec<&str> = action_option.command.split_whitespace().collect();

    //let envs_vec = crate::parse_envs_input(&action_option.env);
    //let vols_vec = crate::parse_volumes_input(&action_option.volume);

    let rand_string: String = thread_rng()
        .sample_iter(&Alphanumeric)
        .map(char::from)
        .take(7)
        .collect();

    let container_name = format!(
        "{}",
        orbital_agent::generate_unique_build_id(
            "test-org",
            "test-repo",
            "test-hash",
            &format!("{}", rand_string),
        )
    );

    let build_container_spec = OrbitalContainerSpec {
        name: Some(container_name),
        image: action_option.image.clone(),
        command: action_option.command.split_whitespace().collect(),
        env_vars: crate::parse_envs_input(&action_option.env),
        volumes: crate::parse_volumes_input(&action_option.volume),
        timeout: Some(Duration::from_secs(60 * 30)), // 30 min
    };

    match docker::container_create(build_container_spec) {
        Ok(container_id) => {
            println!("{}", container_id);
            Ok(())
        }
        Err(_) => Err(SubcommandError::new(&format!(
            "Could not pull image {:?}",
            &action_option.image
        ))
        .into()),
    }
}

extern crate structopt;
use std::str::FromStr;
use structopt::StructOpt;

use container_builder::docker;
use log::debug;

use crate::{GlobalOption, SubcommandError};

/// Local options for the Docker developer subcommand
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Docker image. If no tag provided, :latest will be assumed
    #[structopt(long)]
    image: Option<String>,

    /// ID of an existing Docker container
    #[structopt(long)]
    container_id: Option<String>,

    /// String command to execute in container. Will naively split on whitespace.
    #[structopt(long)]
    command: Option<String>,

    /// Add env vars to build. Comma-separated with no spaces. ex. "key1=var1,key2=var2"
    #[structopt(long, short)]
    env: Option<String>,

    /// Add volume mapping from host to container. Comma-separated with no spaces. ex. "/host/path1:/container/path1,/host/path2:/container/path2"
    #[structopt(long, short)]
    volume: Option<String>,

    /// Pull, Create, Start, Stop, Exec
    action: Action,
}

/// Represents the docker cli actions supported by Docker api wrapper
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub enum Action {
    /// Wrapped call for `docker pull`
    Pull,
    /// Wrapped call for `docker create`
    Create,
    /// Wrapped call for `docker start`
    Start,
    /// Wrapped call for `docker stop`
    Stop,
    /// Wrapped call for `docker exec`
    Exec,
}

/// Naive parse of  a string to one of the supported Docker api actions
impl FromStr for Action {
    type Err = String;
    fn from_str(action: &str) -> Result<Self, Self::Err> {
        match action.to_ascii_lowercase().as_ref() {
            "pull" => Ok(Action::Pull),
            "create" => Ok(Action::Create),
            "start" => Ok(Action::Start),
            "stop" => Ok(Action::Stop),
            "exec" => Ok(Action::Exec),
            _ => Err("Invalid action".to_string()),
        }
    }
}

/// Intended use for Orb developers to use the internal Docker api wrapper calls outside of a build context.
/// # Pull
/// Pull an image through the Docker api
/// Expects `--image` to be provided, and uses the host Docker engine to pull the image.
/// If no build tag is provided, then `:latest` will be assumed.
///
/// The equivalent `docker` command is `docker pull <image>`
/// # Create
/// Create a container running a given command of a given image.
/// Expects `--image` and `--command` to be provided. Splits the command on whitespace, so beware of complex one-liners.
/// By default, the host's docker socket is mounted into the container as `/var/run/docker.sock:/var/run/docker.sock`
/// Returns a container id.
///
/// The equivalent `docker` command is `docker create <image> <command> [--env list] [--volume list]`
/// # Start
/// Starts a container from a given container id
/// Expects `--container-id`
///
/// The equivalent `docker` command is `docker start <container id>`
/// # Stop
/// Stops a container from a given container id
/// Expects `--container-id`
///
/// The equivalent `docker` command is `docker stop <container id>`
/// # Exec
/// Executes a given command into a running container by id
/// Expects `--image` and `--command` to be provided. Splits the command on whitespace, so beware of complex one-liners.
///
/// The equivalent `docker` command is `docker exec <container id> <command>`
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    match local_option.action {
        Action::Pull => {
            match docker::container_pull(
                local_option
                    .image
                    .clone()
                    .expect("No image provided")
                    .as_str(),
            ) {
                Ok(_) => return Ok(()),
                Err(_) => {
                    return Err(SubcommandError::new(&format!(
                        "Could not pull image {:?}",
                        &local_option.image
                    )))
                }
            };
        }
        Action::Create => {
            let unwrapped_command = local_option.command.clone().expect("No command provided");

            // FIXME
            // This is going to be a stupid parsed command on whitespace only.
            // Embedded commands with quotes, $(), or backtics not expected to work with this parsing
            let command_vec_slice: Vec<&str> = unwrapped_command.split_whitespace().collect();

            let envs_vec = super::parse_envs_input(&local_option.env);
            let vols_vec = super::parse_volumes_input(&local_option.volume);

            match docker::container_create(
                local_option
                    .image
                    .clone()
                    .expect("No image provided")
                    .as_str(),
                command_vec_slice,
                envs_vec,
                vols_vec,
            ) {
                Ok(container_id) => {
                    println!("{}", container_id);
                    return Ok(());
                }
                Err(_) => {
                    return Err(SubcommandError::new(&format!(
                        "Could not pull image {:?}",
                        &local_option.image
                    )))
                }
            };
        }

        Action::Start => {
            debug!("Starting container");
            let container_id = local_option
                .container_id
                .clone()
                .expect("No container id provided");
            match docker::container_start(&container_id) {
                Ok(container_id) => container_id,
                Err(_) => {
                    return Err(SubcommandError::new(&format!(
                        "Could not start Docker container id  {}",
                        container_id
                    )))
                }
            }
            Ok(())
        }

        Action::Stop => {
            debug!("Stopping container");
            let container_id = local_option
                .container_id
                .clone()
                .expect("No container id provided");
            match docker::container_stop(&container_id) {
                Ok(container_id) => container_id,
                Err(_) => {
                    return Err(SubcommandError::new(&format!(
                        "Could not stop Docker container id  {}",
                        container_id
                    )))
                }
            }
            Ok(())
        }
        Action::Exec => {
            debug!("Exec'ing commands into container");
            let container_command = &local_option.command.clone().expect("No command provided");
            let container_id = local_option
                .container_id
                .clone()
                .expect("No container id provided");

            // FIXME
            // This is going to be a stupid parsed command on whitespace only.
            // Embedded commands with quotes, $(), or backtics not expected to work with this parsing
            let command_vec_slice: Vec<&str> = container_command.split_whitespace().collect();

            match docker::container_exec(container_id.as_ref(), command_vec_slice.clone()) {
                Ok(output) => {
                    debug!("Command: {:?}", &command_vec_slice);
                    debug!("Output: {:?}", &output);
                    output
                }
                Err(_) => {
                    return Err(SubcommandError::new(&format!(
                        "Could not exec into container id {}",
                        &container_id
                    )))
                }
            }
            Ok(())
        }
    }
}

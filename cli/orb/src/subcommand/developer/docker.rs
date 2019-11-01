extern crate structopt;
use std::str::FromStr;
use structopt::StructOpt;

use container_builder::docker;
use log::debug;

use crate::{GlobalOption, SubcommandError};

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Docker image
    #[structopt(long)]
    image: Option<String>,

    /// ID of an existing Docker container
    #[structopt(long)]
    container_id: Option<String>,

    /// command
    #[structopt(long)]
    command: Option<String>,

    /// Add env vars to build. Comma-separated with no spaces. ex. "key1=var1,key2=var2"
    #[structopt(long, short)]
    env: Option<String>,

    /// Add volume mapping from host to container. Comma-separated with no spaces. ex. "/host/path1:/container/path1,/host/path2:/container/path2"
    #[structopt(long, short)]
    volume: Option<String>,

    /// Pull, Create
    action: Action,
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub enum Action {
    Pull,
    Create,
    Start,
    Stop,
    Exec,
}

impl FromStr for Action {
    type Err = String;
    fn from_str(action: &str) -> Result<Self, Self::Err> {
        match action {
            "pull" => Ok(Action::Pull),
            "create" => Ok(Action::Create),
            _ => Err("Invalid action".to_string()),
        }
    }
}

pub fn subcommand_handler(
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

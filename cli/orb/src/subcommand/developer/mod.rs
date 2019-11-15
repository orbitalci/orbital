use structopt::StructOpt;

use crate::{GlobalOption, SubcommandError};

use log::debug;
use std::env;
use std::io;

use agent_runtime;
/// Generate command line shell completions
pub mod completion;
/// Access into internal Docker wrapper library
pub mod docker;
/// Access into internal git library
pub mod git;
/// Experience the remote build workflows locally
pub mod local_build;
/// Validate `orb.yml` config files
pub mod validate;

/// Subcommands for `orb developer`
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub enum DeveloperType {
    /// Test git repo metadata parser
    Git(git::SubcommandOption),
    /// Test the docker driver
    Docker(docker::SubcommandOption),
    /// Test running builds
    Build(local_build::SubcommandOption),
    /// Test the config file parsers
    Validate(validate::SubcommandOption),
    /// Generate shell completions script for orb command
    Completion(completion::SubcommandOption),
}

/// Returns a `String` of the current working directory.
pub fn get_current_workdir() -> String {
    let path = match env::current_dir() {
        Ok(d) => format!("{}", d.display()),
        Err(_) => String::from("."),
    };

    debug!("Current workdir on host: {}", &path);
    path
}

/// Wrapper function for `kv_csv_parser` to specifically handle env vars for `shiplift`
pub fn parse_envs_input(user_input: &Option<String>) -> Option<Vec<&str>> {
    let envs = kv_csv_parser(user_input);
    debug!("Env vars to set: {:?}", envs);
    envs
}

/// Wrapper function for `kv_csv_parser` to specifically handle volume mounts for `shiplift`
/// Automatically add in the docker socket as defined by `agent_runtime::DOCKER_SOCKET_VOLMAP`. If we don't pass in any other volumes
///
/// For now, also assume passing in the current working directory as well
pub fn parse_volumes_input(user_input: &Option<String>) -> Option<Vec<&str>> {
    let vols = match kv_csv_parser(user_input) {
        Some(v) => {
            let mut new_vec: Vec<&str> = Vec::new();
            new_vec.push(agent_runtime::DOCKER_SOCKET_VOLMAP);
            new_vec.extend(v.clone());
            Some(new_vec)
        }
        None => {
            let mut new_vec: Vec<&str> = Vec::new();
            new_vec.push(agent_runtime::DOCKER_SOCKET_VOLMAP);

            // There's got to be a better way to handle this...
            // https://stackoverflow.com/a/30527289/1672638
            new_vec.push(Box::leak(
                format!(
                    "{}:{}",
                    get_current_workdir(),
                    agent_runtime::ORBITAL_CONTAINER_WORKDIR,
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

/// Subcommand router for `orb developer`
pub async fn subcommand_handler(
    global_option: GlobalOption,
    dev_subcommand: DeveloperType,
) -> Result<(), SubcommandError> {
    match dev_subcommand {
        DeveloperType::Git(sub_option) => git::subcommand_handler(global_option, sub_option).await,
        DeveloperType::Docker(sub_option) => {
            docker::subcommand_handler(global_option, sub_option).await
        }
        DeveloperType::Build(sub_option) => {
            local_build::subcommand_handler(global_option, sub_option).await
        }
        DeveloperType::Validate(sub_option) => {
            validate::subcommand_handler(global_option, sub_option).await
        }
        DeveloperType::Completion(shell) => {
            GlobalOption::clap().gen_completions_to(
                env!("CARGO_PKG_NAME"),
                shell.into(),
                &mut io::stdout(),
            );
            Ok(())
        }
    }
}

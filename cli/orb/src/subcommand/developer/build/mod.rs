extern crate structopt;
use structopt::StructOpt;

use agent_runtime::docker;
use config_parser::yaml as parser;
use git_meta::git_info;
use std::path::Path;

use log::debug;

use crate::{GlobalOption, SubcommandError};
use std::path::PathBuf;

/// Local options for customizing local docker build with orb
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long, parse(from_os_str), env = "PWD")]
    path: PathBuf,

    /// Add env vars to build. Comma-separated with no spaces. ex. "key1=var1,key2=var2"
    #[structopt(long, short)]
    env: Option<String>,

    /// Add volume mapping from host to container. Comma-separated with no spaces. ex. "/host/path1:/container/path1,/host/path2:/container/path2"
    #[structopt(long, short)]
    volume: Option<String>,

    /// Use the specified local branch
    #[structopt(long)]
    branch: Option<String>,

    /// Use the specified commit hash
    #[structopt(long)]
    hash: Option<String>,
}

/// If `--path` not given, expects current working directory, and parses for git metadata
/// Reads `orb.yml` from path. Pulls, creates and starts a container from the specified yaml.
/// Loops over the command list and executes commands into the container
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    // Read options and validate against git repo
    // Read orb.yml

    // If a path isn't given, then use the current working directory based on env var PWD
    let path = &local_option.path;

    let envs_vec = crate::parse_envs_input(&local_option.env);
    let vols_vec = crate::parse_volumes_input(&local_option.volume);

    debug!(
        "Git info at path ({:?}): {:?}",
        &path,
        git_info::get_git_info_from_path(path, &None, &None)
    );

    // TODO: Will want ability to pass in any yaml.
    // TODO: Also handle file being named orb.yaml
    // Look for a file named orb.yml
    debug!("Loading orb.yml from path {:?}", &path);
    let config = parser::load_orb_yaml(Path::new(&format!("{}/{}", &path.display(), "orb.yml")))?;

    debug!("Pulling container: {:?}", config.image.clone());
    match docker::container_pull(config.image.as_str()) {
        Ok(ok) => ok, // The successful result doesn't matter
        Err(_) => {
            return Err(SubcommandError::new(&format!(
                "Could not pull image {}",
                &config.image
            )))
        }
    };

    // Create a new container
    debug!("Creating container");
    let default_command_w_timeout = vec!["sleep", "2h"];
    let container_id = match docker::container_create(
        config.image.as_str(),
        default_command_w_timeout,
        envs_vec,
        vols_vec,
    ) {
        Ok(container_id) => container_id,
        Err(_) => {
            return Err(SubcommandError::new(&format!(
                "Could not create image {}",
                &config.image
            )))
        }
    };

    // Start the new container

    match docker::container_start(&container_id) {
        Ok(container_id) => container_id,
        Err(_) => {
            return Err(SubcommandError::new(&format!(
                "Could not start image {}",
                &config.image
            )))
        }
    }

    // TODO: Make sure tests try to exec w/o starting the container
    // Exec into the new container
    debug!("Sending commands into container");
    for command in config.command.iter() {
        // Build the exec string
        let wrapped_command = format!("{} | tee -a /proc/1/fd/1", &command);

        let container_command = vec!["/bin/sh", "-c", wrapped_command.as_ref()];

        match docker::container_exec(container_id.as_ref(), container_command.clone()) {
            Ok(output) => {
                debug!("Command: {:?}", &command);
                debug!("Output: {:?}", &output);
                output
            }
            Err(_) => {
                return Err(SubcommandError::new(&format!(
                    "Could not create image {}",
                    &config.image
                )))
            }
        }
    }

    Ok(())
}

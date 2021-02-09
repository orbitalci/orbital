use color_eyre::eyre::Result;
use structopt::StructOpt;

use config_parser::yaml as parser;
use git_meta::GitRepo;
use orbital_agent::generate_unique_build_id;
use orbital_exec_runtime::{
    self, docker, docker::OrbitalContainerSpec, parse_envs_input, parse_volumes_input,
};
use std::path::Path;

use log::debug;

use crate::{GlobalOption, SubcommandError};
use std::path::PathBuf;

use rand::distributions::Alphanumeric;
use rand::{thread_rng, Rng};
use std::time::Duration;

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
) -> Result<()> {
    // Read options and validate against git repo
    // Read orb.yml

    // If a path isn't given, then use the current working directory based on env var PWD
    let path = &local_option.path;
    let git_info = GitRepo::open(path.to_path_buf(), None, None).expect("Could not parse repo");

    debug!("Git info at path ({:?}): {:?}", &path, &git_info);

    // TODO: Will want ability to pass in any yaml.
    // TODO: Also handle file being named orb.yaml
    // Look for a file named orb.yml
    debug!("Loading orb.yml from path {:?}", &path);
    //let config = parser::load_orb_yaml(Path::new(&format!("{}/{}", &path.display(), "orb.yml")))?;
    let config =
        parser::load_orb_yaml(Path::new(&format!("{}/{}", &path.display(), "orb.yml"))).unwrap();

    let rand_string: String = thread_rng()
        .sample_iter(&Alphanumeric)
        .map(char::from)
        .take(7)
        .collect();
    let default_command_w_timeout = vec!["sleep", "2h"];
    let build_container_spec = OrbitalContainerSpec {
        name: Some(generate_unique_build_id(
            "dev-org",
            &git_info.url.name,
            &git_info.head.unwrap().id,
            &format!("{}", rand_string),
        )),
        image: config.image,
        command: default_command_w_timeout,

        // TODO: Inject the dynamic build env vars here
        env_vars: parse_envs_input(&None),
        volumes: parse_volumes_input(&None),
        timeout: Some(Duration::from_secs(60 * 30)),
    };

    debug!(
        "Pulling container: {:?}",
        build_container_spec.image.clone()
    );
    match docker::container_pull(&build_container_spec.image.clone()) {
        Ok(ok) => ok, // The successful result doesn't matter
        Err(_) => {
            return Err(SubcommandError::new(&format!(
                "Could not pull image {}",
                &build_container_spec.image
            ))
            .into())
        }
    };

    // Create a new container
    debug!("Creating container");
    let container_id = match docker::container_create(build_container_spec.clone()) {
        Ok(container_id) => container_id,
        Err(_) => {
            return Err(SubcommandError::new(&format!(
                "Could not create image {}",
                &build_container_spec.image
            ))
            .into())
        }
    };

    // Start the new container

    match docker::container_start(&container_id) {
        Ok(container_id) => container_id,
        Err(_) => {
            return Err(SubcommandError::new(&format!(
                "Could not start image {}",
                &build_container_spec.image
            ))
            .into())
        }
    }

    // TODO: Loop on stages
    // TODO: Make sure tests try to exec w/o starting the container
    // Exec into the new container
    debug!("Sending commands into container");
    let mut exec_output: Vec<String> = Vec::new();
    for command in config.stages[0].command.clone().iter() {
        // Build the exec string
        let wrapped_command = format!("{} | tee -a /proc/1/fd/1", &command);

        let container_command = vec!["/bin/sh", "-c", wrapped_command.as_ref()];

        match docker::container_exec(container_id.as_ref(), container_command.clone()) {
            Ok(output) => {
                debug!("Command: {:?}", &command);
                debug!("Output: {:?}", &output);
                &mut exec_output.extend(output);
                ()
            }
            Err(_) => {
                return Err(SubcommandError::new(&format!(
                    "Could not create image {}",
                    &build_container_spec.image
                ))
                .into())
            }
        }
    }

    Ok(())
}

use color_eyre::eyre::Result;
use structopt::StructOpt;

use crate::subcommand::GlobalOption;

pub mod create;
pub mod exec;
pub mod pull;
pub mod start;
pub mod stop;

/// Local options for the Docker developer subcommand
#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Pull, Create, Start, Stop, Exec
    #[structopt(subcommand)]
    action: Action,
}

/// Represents the docker cli actions supported by Docker api wrapper
#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub enum Action {
    /// Wrapped call for `docker create`
    Create(create::ActionOption),
    /// Wrapped call for `docker exec`
    Exec(exec::ActionOption),
    /// Wrapped call for `docker pull`
    Pull(pull::ActionOption),
    /// Wrapped call for `docker start`
    Start(start::ActionOption),
    /// Wrapped call for `docker stop`
    Stop(stop::ActionOption),
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
    global_option: GlobalOption,
    local_option: SubcommandOption,
) -> Result<()> {
    match local_option.clone().action {
        Action::Pull(action_option) => {
            pull::action_handler(global_option, local_option, action_option).await
        }
        Action::Create(action_option) => {
            create::action_handler(global_option, local_option, action_option).await
        }

        Action::Start(action_option) => {
            start::action_handler(global_option, local_option, action_option).await
        }

        Action::Stop(action_option) => {
            stop::action_handler(global_option, local_option, action_option).await
        }

        Action::Exec(action_option) => {
            exec::action_handler(global_option, local_option, action_option).await
        }
    }
}

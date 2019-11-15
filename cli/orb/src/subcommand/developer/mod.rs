use structopt::StructOpt;

use crate::{GlobalOption, SubcommandError};

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

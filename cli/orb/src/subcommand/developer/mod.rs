use color_eyre::eyre::Result;
use structopt::StructOpt;

use crate::{GlobalOption, SubcommandContext};

use std::io;

/// Experience the remote build workflows locally
pub mod build;
/// Generate command line shell completions
pub mod completion;
/// Access into internal Docker wrapper library
pub mod docker;
/// Access into internal git library
pub mod git;
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
    Build(build::SubcommandOption),
    /// Test the config file parsers
    Validate(validate::SubcommandOption),
    /// Generate shell completions script for orb command
    Completion(completion::SubcommandOption),
}

/// Subcommand router for `orb developer`
pub async fn subcommand_handler(
    global_option: GlobalOption,
    dev_subcommand: DeveloperType,
) -> Result<()> {
    match dev_subcommand {
        DeveloperType::Git(sub_option) => git::subcommand_handler(global_option, sub_option).await,
        DeveloperType::Docker(sub_option) => {
            docker::subcommand_handler(global_option, sub_option).await
        }
        DeveloperType::Build(sub_option) => {
            build::subcommand_handler(global_option, sub_option).await
        }
        DeveloperType::Validate(sub_option) => {
            validate::subcommand_handler(global_option, sub_option).await
        }
        DeveloperType::Completion(shell) => {
            SubcommandContext::clap().gen_completions_to("orb", shell.into(), &mut io::stdout());
            Ok(())
        }
    }
}

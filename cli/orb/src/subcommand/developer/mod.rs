use structopt::StructOpt;

use crate::{GlobalOption, SubcommandError};

pub mod docker;
pub mod git;
pub mod local_build;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub enum DeveloperType {
    Git(git::SubcommandOption),
    Docker(docker::SubcommandOption),
    Build(local_build::SubcommandOption),
}

pub fn subcommand_handler(
    global_option: GlobalOption,
    dev_subcommand: DeveloperType,
) -> Result<(), SubcommandError> {
    match dev_subcommand {
        DeveloperType::Git(sub_option) => git::subcommand_handler(global_option, sub_option),
        DeveloperType::Docker(sub_option) => docker::subcommand_handler(global_option, sub_option),
        DeveloperType::Build(sub_option) => local_build::subcommand_handler(global_option, sub_option),
    }
}

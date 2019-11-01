extern crate structopt;
use structopt::StructOpt;

use git_meta::git_info;

use crate::{GlobalOption, SubcommandError};

use log::debug;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,
}

pub fn subcommand_handler(
    _global_option: GlobalOption,
    local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    if let Some(path) = local_option.path {
        debug!(
            "Git path: {:?}",
            git_info::get_git_info_from_path(&path, &None, &None)
        );
    }

    Ok(())
}

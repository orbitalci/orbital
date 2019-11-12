extern crate structopt;
use structopt::StructOpt;

use git_meta::git_info;

use crate::{GlobalOption, SubcommandError};

use log::debug;

/// Local options for customizing git library call
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,
}

/// Expects `--path`. Attempts to open directory and parse repo for git metadata and prints to stdout
pub async fn subcommand_handler(
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

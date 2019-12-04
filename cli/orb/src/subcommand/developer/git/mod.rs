extern crate structopt;
use structopt::StructOpt;

use crate::{GlobalOption, SubcommandError};

pub mod clone;
pub mod info;

//use log::debug;

/// Local options for customizing git library call
#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,

    /// info, clone
    #[structopt(subcommand)]
    action: Action,
}

/// Represents the main git workflows
#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub enum Action {
    /// Parse a path for its git remote info
    Info(info::ActionOption),
    /// Clone a git repo
    Clone(clone::ActionOption),
}

/// Expects `--path`. Attempts to open directory and parse repo for git metadata and prints to stdout
pub async fn subcommand_handler(
    global_option: GlobalOption,
    local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    match local_option.clone().action {
        Action::Info(action_option) => {
            info::action_handler(global_option, local_option, action_option).await
        }
        Action::Clone(action_option) => {
            clone::action_handler(global_option, local_option, action_option).await
        }
    }
}

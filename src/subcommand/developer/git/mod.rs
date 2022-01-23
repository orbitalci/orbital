use color_eyre::eyre::Result;
use structopt::StructOpt;

use crate::subcommand::GlobalOption;

pub mod clone;
pub mod info;
pub mod poll;

//use std::path::PathBuf;
//use log::debug;

/// Local options for customizing git library call
#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    // TODO: Handle these flags

    ///// Path to local repo. Defaults to current working directory
    //#[structopt(long, parse(from_os_str), env = "PWD")]
    //path: PathBuf,
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
    /// Simulate a repo poll - Shallow clone a git repo, then list remote references.
    Poll(poll::ActionOption),
}

/// Expects `--path`. Attempts to open directory and parse repo for git metadata and prints to stdout
pub async fn subcommand_handler(
    global_option: GlobalOption,
    local_option: SubcommandOption,
) -> Result<()> {
    match local_option.clone().action {
        Action::Info(action_option) => {
            info::action_handler(global_option, local_option, action_option).await
        }
        Action::Clone(action_option) => {
            clone::action_handler(global_option, local_option, action_option).await
        }
        Action::Poll(action_option) => {
            poll::action_handler(global_option, local_option, action_option).await
        }
    }
}

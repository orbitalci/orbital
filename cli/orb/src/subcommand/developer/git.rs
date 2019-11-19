extern crate structopt;
use std::fs;
use std::str::FromStr;
use structopt::StructOpt;

use git_meta::{git_info, GitCredentials};

use agent_runtime::build_engine;

use crate::{GlobalOption, SubcommandError};

use log::debug;

/// Local options for customizing git library call
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,

    /// info, clone
    action: Action,
}

/// Represents the main git workflows
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub enum Action {
    /// Parse a path for its git remote info
    Info,
    /// Clone a git repo
    Clone,
}

/// Naive parse of a string to one of the supported git actions
impl FromStr for Action {
    type Err = String;
    fn from_str(action: &str) -> Result<Self, Self::Err> {
        match action.to_ascii_lowercase().as_ref() {
            "info" => Ok(Action::Info),
            "clone" => Ok(Action::Clone),
            _ => Err("Invalid action".to_string()),
        }
    }
}

/// Expects `--path`. Attempts to open directory and parse repo for git metadata and prints to stdout
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    match local_option.action {
        // Print the git repo info
        Action::Info => {
            println!(
                "{:?}",
                git_info::get_git_info_from_path(
                    crate::get_current_workdir().as_ref(),
                    &None,
                    &None
                )
            );
        }
        Action::Clone => {
            let temp_dir = build_engine::clone_repo(
                "https://github.com/alexcrichton/git2-rs",
                GitCredentials::Public,
            )?;

            let paths = fs::read_dir(&temp_dir.as_path()).unwrap();
            for path in paths {
                println!("Name: {}", path.unwrap().path().display())
            }
        }
    }

    //if let Some(path) = local_option.path {
    //    debug!(
    //        "Git path: {:?}",
    //        git_info::get_git_info_from_path(&path, &None, &None)
    //    );
    //}

    Ok(())
}

use crate::{developer::git::SubcommandOption, GlobalOption, SubcommandError};
//use log::debug;
use agent_runtime::build_engine;
use git_meta::GitCredentials;
use std::fs;
use structopt::StructOpt;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    _action_option: ActionOption,
) -> Result<(), SubcommandError> {
    let temp_dir = build_engine::clone_repo(
        "https://github.com/alexcrichton/git2-rs",
        "master",
        GitCredentials::Public,
    )?;

    let paths = fs::read_dir(&temp_dir.as_path()).unwrap();
    for path in paths {
        println!("Name: {}", path.unwrap().path().display())
    }

    Ok(())
}

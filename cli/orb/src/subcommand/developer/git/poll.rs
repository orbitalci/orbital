use crate::{developer::git::SubcommandOption, GlobalOption};
//use log::debug;
use anyhow::Result;
use git_meta::{self, GitCredentials};
use mktemp::Temp;
use std::fs;
use structopt::StructOpt;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    #[structopt(short, long)]
    priv_key: Option<String>,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<()> {
    let temp_dir = Temp::new_dir().expect("Unable to create test clone dir");

    if let Some(key) = action_option.priv_key {
    } else {
        let _res = git_meta::clone::shell_shallow_clone(
            "https://github.com/alexcrichton/git2-rs",
            None,
            GitCredentials::Public,
            &temp_dir.as_path(),
            //)?;
        )
        .unwrap();
    }

    // do ls-remote to temp_dir
    println!(
        "{:?}",
        git_meta::git_info::list_remote_branch_heads(&temp_dir.as_path())
    );

    Ok(())
}

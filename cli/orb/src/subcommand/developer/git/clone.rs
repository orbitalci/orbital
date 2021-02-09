use crate::{developer::git::SubcommandOption, GlobalOption};
//use log::debug;
use anyhow::Result;
use mktemp::Temp;
use orbital_agent::build_engine;
use std::fs;
use structopt::StructOpt;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    #[structopt(short, long)]
    branch: Option<String>,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<()> {
    let temp_dir = Temp::new_dir().expect("Unable to create test clone dir");

    let _res = build_engine::clone_repo(
        "https://github.com/alexcrichton/git2-rs",
        action_option.branch.as_deref(),
        None,
        &temp_dir.as_path(),
        //)?;
    )
    .unwrap();

    let paths = fs::read_dir(&temp_dir.as_path()).unwrap();
    for path in paths {
        println!("Name: {}", path.unwrap().path().display())
    }

    Ok(())
}

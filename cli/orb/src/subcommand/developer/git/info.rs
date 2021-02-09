use crate::{developer::git::SubcommandOption, GlobalOption};
use color_eyre::eyre::Result;
use git_meta::GitRepo;
use std::path::PathBuf;
use structopt::StructOpt;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    /// Repo path
    #[structopt(parse(from_os_str), env = "PWD")]
    path: PathBuf,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<()> {
    println!("{:?}", GitRepo::open(action_option.path, None, None));

    Ok(())
}

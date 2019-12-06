use crate::{developer::git::SubcommandOption, GlobalOption, SubcommandError};
//use log::debug;
use git_meta::git_info;
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
) -> Result<(), SubcommandError> {
    println!(
        "{:?}",
        git_info::get_git_info_from_path(&action_option.path, &None, &None)
    );

    Ok(())
}

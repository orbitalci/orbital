use crate::subcommand::{developer::git::SubcommandOption, GlobalOption};
//use log::debug;
use color_eyre::eyre::Result;
use git_meta::{self, GitRepo};
use mktemp::Temp;
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

    if let Some(_key) = action_option.priv_key {
    } else {
        let _git2repo = GitRepo::new("https://github.com/alexcrichton/git2-rs")
            .expect("Unable to create GitRepo")
            .git_clone_shallow(temp_dir.as_path())
            .expect("Unable to clone repo");
    }

    let repo = GitRepo::open(temp_dir.as_path().into(), None, None)
        .expect("Unable to open repo directory");
    // do ls-remote to temp_dir
    println!("{:?}", repo.get_remote_branch_head_refs(None));

    Ok(())
}

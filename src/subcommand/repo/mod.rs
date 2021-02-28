use color_eyre::eyre::Result;
use structopt::StructOpt;

use crate::subcommand::GlobalOption;
use std::path::PathBuf;

pub mod add;
pub mod get;
pub mod list;
pub mod remove;
pub mod update;

/// Local options for customizing repo request
#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long, parse(from_os_str), env = "PWD")]
    path: PathBuf,

    #[structopt(subcommand)]
    pub action: Action,
}

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub enum Action {
    Add(add::ActionOption),
    Get(get::ActionOption),
    Update(update::ActionOption),
    #[structopt(alias = "rm")]
    Remove(remove::ActionOption),
    #[structopt(alias = "ls")]
    List(list::ActionOption),
    // Enable,
    // Disable,
}

/// *Not yet implemented* Backend calls for managing repo resources
pub async fn subcommand_handler(
    global_option: GlobalOption,
    local_option: SubcommandOption,
) -> Result<()> {
    match local_option.clone().action {
        Action::Add(action_option) => {
            add::action_handler(global_option, local_option, action_option).await
        }
        Action::Get(action_option) => {
            get::action_handler(global_option, local_option, action_option).await
        }
        Action::Update(action_option) => {
            update::action_handler(global_option, local_option, action_option).await
        }
        Action::Remove(action_option) => {
            remove::action_handler(global_option, local_option, action_option).await
        }
        Action::List(action_option) => {
            list::action_handler(global_option, local_option, action_option).await
        }
    }
}

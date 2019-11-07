extern crate structopt;
use structopt::StructOpt;

use crate::{GlobalOption, SubcommandError};

/// Local options for customizing polling request
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,
}

/// *Not yet implemented* Backend calls for managing polling resources
pub fn subcommand_handler(
    _global_option: GlobalOption,
    _local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    Err(SubcommandError::new("Not yet implemented"))
}

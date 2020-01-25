use anyhow::Result;
use structopt::StructOpt;

use crate::GlobalOption;

/// Start an Orb server
pub mod start;

/// Subcommands for `orb operator`
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub enum OperatorType {
    Start(start::SubcommandOption),
}

/// Subcommand router for `orb operator`
pub async fn subcommand_handler(
    global_option: GlobalOption,
    ops_subcommand: OperatorType,
) -> Result<()> {
    match ops_subcommand {
        OperatorType::Start(sub_option) => {
            start::subcommand_handler(global_option, sub_option).await
        }
    }
}

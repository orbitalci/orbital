use structopt::StructOpt;

use crate::{GlobalOption, SubcommandError};

pub mod start;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub enum OperatorType {
    Start(start::SubcommandOption),
}

pub fn subcommand_handler(
    global_option: GlobalOption,
    ops_subcommand: OperatorType,
) -> Result<(), SubcommandError> {
    match ops_subcommand {
        OperatorType::Start(sub_option) => start::subcommand_handler(global_option, sub_option),
    }
}

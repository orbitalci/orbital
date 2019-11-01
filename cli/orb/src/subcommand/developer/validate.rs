extern crate structopt;
use structopt::StructOpt;

use config_parser::parser;

use crate::{GlobalOption, SubcommandError};

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to orb config file. Defaults to current working directory
    #[structopt(long, short)]
    file: Option<String>,
}

// TODO: We want to return the config
pub fn subcommand_handler(
    _global_option: GlobalOption,
    local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    if let Some(yaml_file) = local_option.file {
        match parser::load_orb_yaml(yaml_file) {
            Ok(_c) => Ok(()),
            Err(_e) => Err(SubcommandError::new("Config file failed validation")),
        }
    } else {
        Err(SubcommandError::new("No config file specified"))
    }
}

use anyhow::Result;
use structopt::StructOpt;

use config_parser::yaml as parser;
use std::path::PathBuf;

use crate::{GlobalOption, SubcommandError};

/// Local options for customizing validation of an Orbital config file
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to orb config file. Defaults to current working directory
    #[structopt(parse(from_os_str), env = "PWD")]
    file: PathBuf,
}

// TODO: We want to return the config
/// Validate the config by loading it. Serde-yaml will error out if there are syntax issues.
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    local_option: SubcommandOption,
) -> Result<()> {
    match parser::load_orb_yaml(&local_option.file.as_path()) {
        Ok(c) => {
            //println!("Full config:\n{:?}", c);

            //let mut marker = (0, 0);

            //loop {
            //    let stage_index = marker.0;
            //    let command_index = marker.1;

            //    println!(
            //        "Stage index:{} Command index:{}",
            //        stage_index, command_index
            //    );

            //    println!("{:?}", c.stages[stage_index].command[command_index]);

            //    if c.stages
            //        .get(stage_index)
            //        .unwrap()
            //        .command
            //        .get(command_index + 1)
            //        .is_some()
            //    {
            //        marker = (stage_index, command_index + 1);
            //    } else if c.stages.get(stage_index + 1).is_some() {
            //        marker = (stage_index + 1, 0);
            //    } else {
            //        break;
            //    }
            //}

            Ok(())
        }
        Err(_e) => Err(SubcommandError::new("Config file failed validation").into()),
    }
}

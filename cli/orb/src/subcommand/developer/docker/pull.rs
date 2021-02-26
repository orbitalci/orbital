use crate::subcommand::{developer::docker::SubcommandOption, GlobalOption, SubcommandError};
use color_eyre::eyre::Result;
use orbital_exec_runtime::docker;
use structopt::StructOpt;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    /// Docker image. If no tag provided, :latest will be assumed
    image: String,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<()> {
    match docker::container_pull(action_option.image.clone().as_str()).await {
        Ok(_) => Ok(()),
        Err(_) => Err(SubcommandError::new(&format!(
            "Could not pull image {:?}",
            &action_option.image
        ))
        .into()),
    }
}

use crate::orbital_utils::exec_runtime::docker;
use crate::subcommand::{developer::docker::SubcommandOption, GlobalOption, SubcommandError};
use color_eyre::eyre::Result;
use tracing::debug;
use structopt::StructOpt;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    /// ID of an existing Docker container
    container_id: String,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<()> {
    debug!("Starting container");
    match docker::container_start(&action_option.container_id).await {
        Ok(_) => {}
        Err(_) => {
            return Err(SubcommandError::new(&format!(
                "Could not start Docker container id  {}",
                &action_option.container_id
            ))
            .into())
        }
    }
    Ok(())
}

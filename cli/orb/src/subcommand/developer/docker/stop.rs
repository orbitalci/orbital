use crate::{developer::docker::SubcommandOption, GlobalOption, SubcommandError};
use agent_runtime::docker;
use log::debug;
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
) -> Result<(), SubcommandError> {
    debug!("Stopping container");
    let container_id = action_option.container_id.clone();
    match docker::container_stop(&container_id) {
        Ok(container_id) => container_id,
        Err(_) => {
            return Err(SubcommandError::new(&format!(
                "Could not stop Docker container id  {}",
                container_id
            )))
        }
    }
    Ok(())
}

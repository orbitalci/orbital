use crate::orbital_utils::exec_runtime::docker;
use crate::subcommand::{developer::docker::SubcommandOption, GlobalOption, SubcommandError};
use color_eyre::eyre::Result;
use structopt::StructOpt;
use tracing::debug;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    /// ID of an existing Docker container
    container_id: String,

    /// String command to execute in container. Will naively split on whitespace.
    command: String,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<()> {
    debug!("Exec'ing commands into container");
    // FIXME
    // This is going to be a stupid parsed command on whitespace only.
    // Embedded commands with quotes, $(), or backtics not expected to work with this parsing
    let command_vec_slice: Vec<&str> = action_option.command.split_whitespace().collect();

    match docker::container_exec(
        action_option.container_id.clone(),
        command_vec_slice.clone(),
    )
    .await
    {
        Ok(mut exec_output) => {
            debug!("Command: {:?}", &command_vec_slice);
            while let Some(output) = exec_output.recv().await {
                print!("Output: {:?}", &output);
            }
        }
        Err(_) => {
            return Err(SubcommandError::new(&format!(
                "Could not exec into container id {}",
                &action_option.container_id.clone()
            ))
            .into())
        }
    }
    Ok(())
}

use crate::{developer::docker::SubcommandOption, GlobalOption, SubcommandError};
use agent_runtime::docker;
//use log::debug;
use structopt::StructOpt;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    /// Add env vars to build. Comma-separated with no spaces. ex. "key1=var1,key2=var2"
    #[structopt(long, short)]
    env: Option<String>,

    /// Add volume mapping from host to container. Comma-separated with no spaces. ex. "/host/path1:/container/path1,/host/path2:/container/path2"
    #[structopt(long, short)]
    volume: Option<String>,

    /// Docker image. If no tag provided, :latest will be assumed
    image: String,

    /// String command to execute in container. Will naively split on whitespace.
    command: String,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<(), SubcommandError> {
    // FIXME
    // This is going to be a stupid parsed command on whitespace only.
    // Embedded commands with quotes, $(), or backtics not expected to work with this parsing
    let command_vec_slice: Vec<&str> = action_option.command.split_whitespace().collect();

    let envs_vec = crate::parse_envs_input(&action_option.env);
    let vols_vec = crate::parse_volumes_input(&action_option.volume);

    match docker::container_create(
        action_option.image.clone().as_str(),
        command_vec_slice,
        envs_vec,
        vols_vec,
    ) {
        Ok(container_id) => {
            println!("{}", container_id);
            return Ok(());
        }
        Err(_) => {
            return Err(SubcommandError::new(&format!(
                "Could not pull image {:?}",
                &action_option.image
            )))
        }
    };
}

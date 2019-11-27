use structopt::StructOpt;

use crate::{repo::SubcommandOption, GlobalOption, SubcommandError};

use orbital_headers::code::{client::CodeServiceClient, GitRepoGetRequest};
use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

use log::debug;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    _action_option: ActionOption,
) -> Result<(), SubcommandError> {
    let mut client = CodeServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let request = Request::new(GitRepoGetRequest {
        ..Default::default()
    });

    debug!("Request for git repo get: {:?}", &request);

    let response = client.git_repo_get(request).await?;
    println!("RESPONSE = {:?}", response);
    Ok(())
}

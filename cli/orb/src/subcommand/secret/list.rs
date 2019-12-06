use structopt::StructOpt;

use crate::{secret::SubcommandOption, GlobalOption, SubcommandError};

use orbital_headers::secret::{client::SecretServiceClient, SecretListRequest};
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
    let mut client = SecretServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let request = Request::new(SecretListRequest {
        ..Default::default()
    });

    debug!("Request for secret list: {:?}", &request);

    let response = client.secret_list(request).await?;
    println!("RESPONSE = {:?}", response);
    Ok(())
}

use structopt::StructOpt;

use crate::{secret::SubcommandOption, GlobalOption, SubcommandError};

use orbital_headers::secret::{client::SecretServiceClient, SecretListRequest};
use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

use log::debug;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    /// Name of Orbital org
    #[structopt(long, env = "ORB_DEFAULT_ORG")]
    org: Option<String>,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<(), SubcommandError> {
    let mut client = SecretServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let request = Request::new(SecretListRequest {
        org: action_option.org.unwrap_or_default(),
        ..Default::default()
    });

    debug!("Request for secret list: {:?}", &request);

    let response = client.secret_list(request).await?;
    println!("RESPONSE = {:?}", response);
    Ok(())
}

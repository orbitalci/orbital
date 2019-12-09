use structopt::StructOpt;

use crate::{org::SubcommandOption, GlobalOption, SubcommandError};

use orbital_headers::organization::client::OrganizationServiceClient;
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
    let mut client =
        OrganizationServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let request = Request::new(());
    debug!("Request for org remove: {:?}", &request);

    let response = client.org_list(request).await?;
    println!("RESPONSE = {:?}", response);
    Ok(())
}

use structopt::StructOpt;

use crate::{org::SubcommandOption, GlobalOption, SubcommandError};

use orbital_headers::organization::{client::OrganizationServiceClient, OrgRemoveRequest};
use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

use log::debug;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    name: String,
    #[structopt(long, short)]
    force: Option<bool>,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<(), SubcommandError> {
    let mut client =
        OrganizationServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let request = Request::new(OrgRemoveRequest {
        name: action_option.name.into(),
        force: action_option
            .force
            .expect("Something went wrong with parsing force flag")
            .into(),
    });
    debug!("Request for org remove: {:?}", &request);

    let response = client.org_remove(request).await?;
    println!("RESPONSE = {:?}", response);
    Ok(())
}

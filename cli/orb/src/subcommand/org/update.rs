use structopt::StructOpt;

use crate::{org::SubcommandOption, GlobalOption, SubcommandError};

use orbital_headers::organization::{client::OrganizationServiceClient, OrgUpdateRequest};
use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

use log::debug;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    name: String,
    update_name: String,
    #[structopt(long, short)]
    active_state: Option<bool>,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<(), SubcommandError> {
    let mut client =
        OrganizationServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let request = Request::new(OrgUpdateRequest {
        name: action_option.name.into(),
        update_name: action_option.update_name.into(),
        active_state: action_option
            .active_state
            .expect("Something went wrong with parsing active state")
            .into(),
    });
    debug!("Request for org update: {:?}", &request);

    let response = client.org_update(request).await?;
    println!("RESPONSE = {:?}", response);
    Ok(())
}

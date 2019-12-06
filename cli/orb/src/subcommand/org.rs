use structopt::StructOpt;

use crate::{GlobalOption, SubcommandError};

use orbital_headers::organization::{client::OrganizationServiceClient, OrgAddRequest};

use orbital_services::ORB_DEFAULT_URI;
use std::path::PathBuf;
use tonic::Request;

/// Local options for customizing org request
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long, parse(from_os_str), env = "PWD")]
    path: PathBuf,
}

/// *Not yet implemented* Backend calls for managing Organization resources
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    _local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    let mut client =
        OrganizationServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let request = Request::new(OrgAddRequest {
        name: "org_name_goes_here".into(),
    });

    let response = client.org_add(request).await?;

    println!("RESPONSE = {:?}", response);

    Ok(())
}

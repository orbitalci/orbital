extern crate structopt;
use structopt::StructOpt;

use crate::{GlobalOption, SubcommandError};

use orbital_headers::organization::{client::OrganizationServiceClient, RepoRegisterRequest};

use crate::ORB_DEFAULT_URI;
use tonic::Request;

/// Local options for customizing repo request
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,
}

/// *Not yet implemented* Backend calls for managing repo resources
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    _local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    let mut client =
        OrganizationServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let request = Request::new(RepoRegisterRequest {
        org: "org_name_goes_here".into(),
        vcs_type: 0,
        uri: "uri_goes_here".into(),
    });

    let response = client.register_repo(request).await?;

    println!("RESPONSE = {:?}", response);

    Ok(())
}

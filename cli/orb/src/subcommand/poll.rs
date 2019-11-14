extern crate structopt;
use structopt::StructOpt;

use crate::{GlobalOption, SubcommandError};

use orbital_headers::organization::{
    client::OrganizationServiceClient,
};

use crate::ORB_DEFAULT_URI;
use tonic::Request;

/// Local options for customizing polling request
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,
}

/// *Not yet implemented* Backend calls for managing polling resources
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    _local_option: SubcommandOption,
) -> Result<(), SubcommandError> {

    // Haven't yet figured out the user-space relationship with polling

    //let mut client =
    //    OrganizationServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    //let request = Request::new(RepoRegisterPollingExpressionRequest {
    //    org: "org_name_goes_here".into(),
    //    account: "account_name_goes_here".into(),
    //    repo: "repo_name_goes_here".into(),
    //    branch: "branch_name_goes_here".into(),
    //    cron_expression: "cron_name_goes_here".into(),
    //});

    //let response = client.poll_repo(request).await?;

    //println!("RESPONSE = {:?}", response);

    Ok(())
}

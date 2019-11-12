extern crate structopt;
use structopt::StructOpt;

use crate::{GlobalOption, SubcommandError};

use orbital_headers::build_metadata::{client::BuildServiceClient, BuildLogRequest};

use crate::ORB_DEFAULT_URI;
use tonic::Request;

/// Local options for customizing summary request
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,
}

// FIXME: Request for summary is not currently served well by proto. How to differeniate from a regular log request?
// Idea: Need a get_summary call. Build id should be Option<u32>, so we can summarize a repo or a specific build
/// *Not yet implemented* Generates request for build summaries
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    _local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    let mut client = BuildServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    // Idea: Index should be Option<u32>
    let request = Request::new(BuildLogRequest {
        org: "org_name_goes_here".into(),
        account: "account_name_goes_here".into(),
        repo: "repo_name_goes_here".into(),
        index: 0,
    });

    let response = client.get_build_logs(request).await?;

    println!("RESPONSE = {:?}", response);

    Ok(())
}

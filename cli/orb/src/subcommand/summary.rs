use structopt::StructOpt;

use crate::GlobalOption;

use orbital_headers::build_meta::{
    build_service_client::BuildServiceClient, BuildSummaryRequest, BuildTarget,
};

use anyhow::Result;
use orbital_services::ORB_DEFAULT_URI;
use std::path::PathBuf;
use tonic::Request;

/// Local options for customizing summary request
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long, parse(from_os_str), env = "PWD")]
    path: PathBuf,

    /// Name of Orbital org
    #[structopt(long, env = "ORB_DEFAULT_ORG")]
    org: Option<String>,
}

// FIXME: Request for summary is not currently served well by proto. How to differeniate from a regular log request?
// Idea: Need a get_summary call. Build id should be Option<u32>, so we can summarize a repo or a specific build
/// *Not yet implemented* Generates request for build summaries
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    _local_option: SubcommandOption,
) -> Result<()> {
    let mut client = BuildServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    // Idea: Index should be Option<u32>
    let request = Request::new(BuildSummaryRequest {
        build: Some(BuildTarget {
            org: "org_name_goes_here".into(),
            ..Default::default()
        }),
        ..Default::default()
    });

    let response = client.build_summary(request).await?;

    println!("RESPONSE = {:?}", response);

    Ok(())
}

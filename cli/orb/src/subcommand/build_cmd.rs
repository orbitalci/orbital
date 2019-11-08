extern crate structopt;
use structopt::StructOpt;

use crate::{GlobalOption, SubcommandError};

use orbital_headers::build_metadata::{client::BuildServiceClient, BuildStartRequest};

/// Local options for customizing build start request
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,
}

/// Generates gRPC `BuildStartRequest` object and connects to *currently hardcoded* gRPC server and sends a request to `BuildService` server.
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    _local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    let mut client = BuildServiceClient::connect("http://127.0.0.1:50051").await?;

    let request = tonic::Request::new(BuildStartRequest {
        remote_uri: "http://1.2.3.4:5678".into(),
        branch: "master".into(),
        commit_ref: "deadbeef".into(),
    });

    let response = client.start_build(request).await?;

    println!("RESPONSE = {:?}", response);

    Ok(())
}

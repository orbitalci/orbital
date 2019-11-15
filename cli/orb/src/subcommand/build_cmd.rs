extern crate structopt;
use structopt::StructOpt;

use crate::{GlobalOption, SubcommandError};

use orbital_headers::build_meta::{client::BuildServiceClient, BuildTarget};

use crate::ORB_DEFAULT_URI;
use tonic::Request;

/// Local options for customizing build start request
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Environment variables to add to build 
    #[structopt(long)]
    envs: Option<String>,
}

/// Generates gRPC `BuildStartRequest` object and connects to *currently hardcoded* gRPC server and sends a request to `BuildService` server.
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    _local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    let mut client = BuildServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

        // Path

        // Read in the git repo
        // uri
        // Git provider
        // Branch
        // Commit
        //  
        // If specified, check if commit is in branch
        // If We're in detatched head (commit not in branch) say so
        //
        // Open the orb.yml
        // Assuming Docker builder... (Stay focused!)
        // Get the docker container image

        // Org - default (Future: How can we cache this client-side?)
        // 


    let request = Request::new(BuildTarget {
        remote_uri: "http://1.2.3.4:5678".into(),
        branch: "master".into(),
        commit_hash: "deadbeef".into(),
        // Docker image
        ..Default::default()
    });

    let response = client.build_start(request).await?;

    println!("RESPONSE = {:?}", response);

    Ok(())
}

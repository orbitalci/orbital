use structopt::StructOpt;

use crate::{GlobalOption, SubcommandError};

use orbital_headers::build_meta::{client::BuildServiceClient, BuildTarget};

use orbital_services::ORB_DEFAULT_URI;
use std::path::PathBuf;
use tonic::Request;

/// Local options for customizing logs request
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long, parse(from_os_str), env = "PWD")]
    path: PathBuf,
}

/// Generates request for logs
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    _local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    let mut client = BuildServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    // Need to figure out how to handle the streaming response

    let request = Request::new(BuildTarget {
        org: "org_name_goes_here".into(),
        git_provider: "default_git_provider".into(),
        git_repo: "default_git_repo".into(),
        remote_uri: "default_remote_uri".into(),
        branch: "default_branch".into(),
        commit_hash: "default_commit_hash".into(),
        envs: "envs".into(),
        id: 0,
    });

    let mut stream = client.build_logs(request).await?.into_inner();

    while let Some(response) = stream.message().await? {
        println!("RESPONSE = {:?}", response);
    }

    Ok(())
}

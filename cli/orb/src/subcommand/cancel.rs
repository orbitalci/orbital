use structopt::StructOpt;

use crate::GlobalOption;

use orbital_headers::build_meta::{build_service_client::BuildServiceClient, BuildTarget};

use anyhow::Result;
use git_meta::GitRepo;
use orbital_services::ORB_DEFAULT_URI;
use std::path::PathBuf;
use tonic::Request;

/// Local options for customizing build cancel request
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long, parse(from_os_str), env = "PWD")]
    path: PathBuf,

    /// Name of Orbital org
    #[structopt(long, env = "ORB_DEFAULT_ORG")]
    org: Option<String>,

    /// Branch name (Default is to choose checked out branch)
    #[structopt(long)]
    branch: Option<String>,

    /// Git commit hash (Default is to choose the remote HEAD commit)
    #[structopt(long)]
    hash: Option<String>,

    /// Build ID (Default is to choose the latest)
    #[structopt(long)]
    id: Option<i32>,
}

/// Generates request for canceling a build in progress
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    local_option: SubcommandOption,
) -> Result<()> {
    let path = &local_option.path;

    let git_context = GitRepo::open(path.to_path_buf(), local_option.branch, local_option.hash)
        .expect("Unable to open GitRepo");

    let mut client = BuildServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let request = Request::new(BuildTarget {
        id: local_option.id.unwrap_or_default(),
        org: local_option.org.expect("Please provide an org name"),
        git_repo: git_context.url.name.clone(),
        remote_uri: git_context.url.trim_auth().to_string(),
        branch: git_context.branch.unwrap_or_default(),
        commit_hash: git_context.head.unwrap().id,
        ..Default::default()
    });

    let response = client.build_stop(request).await?;

    println!("RESPONSE = {:?}", response);

    Ok(())
}

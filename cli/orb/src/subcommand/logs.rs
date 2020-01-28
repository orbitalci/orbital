use structopt::StructOpt;

use crate::GlobalOption;

use orbital_headers::build_meta::{build_service_client::BuildServiceClient, BuildTarget};

use anyhow::Result;
use git_meta::git_info;
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

    /// Name of Orbital org
    #[structopt(long, env = "ORB_DEFAULT_ORG")]
    org: Option<String>,

    /// Git commit hash (Default is to choose the remote HEAD commit)
    #[structopt(long)]
    hash: Option<String>,

    /// Branch name (Default is to choose checked out branch)
    #[structopt(long)]
    branch: Option<String>,

    /// Environment variables to add to build
    #[structopt(long)]
    envs: Option<String>,
}

/// Generates request for logs
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    local_option: SubcommandOption,
) -> Result<()> {
    let mut client = BuildServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    // Path
    let path = &local_option.path;

    let git_context =
        git_info::get_git_info_from_path(path, &local_option.branch, &local_option.hash)?;
    // Need to figure out how to handle the streaming response

    let request = Request::new(BuildTarget {
        org: local_option.org.expect("Please provide an org name"),
        git_repo: git_context.git_url.name,
        remote_uri: git_context.git_url.href,
        //git_provider: "default_git_provider".into(),
        //git_repo: "default_git_repo".into(),
        branch: git_context.branch,
        commit_hash: git_context.commit_id,
        user_envs: local_option.envs.unwrap_or_default(),
        ..Default::default()
    });

    let mut stream = client.build_logs(request).await?.into_inner();

    while let Some(response) = stream.message().await? {
        println!("RESPONSE = {:?}", response);
    }

    Ok(())
}

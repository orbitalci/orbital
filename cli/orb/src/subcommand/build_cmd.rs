use structopt::StructOpt;

use crate::{GlobalOption, SubcommandError};

use orbital_headers::build_meta::{client::BuildServiceClient, BuildTarget};

use config_parser::yaml as parser;
use git_meta::git_info;
use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

use log::debug;
use std::path::{Path, PathBuf};

/// Local options for customizing build start request
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Name of Orbital org
    #[structopt(long, env = "ORB_DEFAULT_ORG")]
    org: Option<String>,

    /// Environment variables to add to build
    #[structopt(long)]
    envs: Option<String>,

    /// Branch name (Default is to choose checked out branch)
    #[structopt(long)]
    branch: Option<String>,

    /// Git commit hash (Default is to choose the remote HEAD commit)
    #[structopt(long)]
    hash: Option<String>,

    /// Path to repo. Defaults to current working directory.
    #[structopt(long, parse(from_os_str), env = "PWD")]
    path: PathBuf,
}

/// Generates gRPC `BuildStartRequest` object and connects to *currently hardcoded* gRPC server and sends a request to `BuildService` server.
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    let mut client = BuildServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    // Path
    let path = &local_option.path;

    // Read in the git repo
    // uri
    // Git provider
    // Branch
    // Commit
    //

    let git_context =
        git_info::get_git_info_from_path(path, &local_option.branch, &local_option.hash)?;
    // If specified, check if commit is in branch
    // If We're in detatched head (commit not in branch) say so
    //
    // Open the orb.yml
    let _config = parser::load_orb_yaml(Path::new(&format!("{}/{}", &path.display(), "orb.yml")))?;
    // Assuming Docker builder... (Stay focused!)
    // Get the docker container image

    // Org - default (Future: How can we cache this client-side?)
    // Validate that the org exists be

    let request = Request::new(BuildTarget {
        org: local_option.org.unwrap_or_default(),
        git_repo: format!("{}/{}", git_context.account, git_context.repo),
        remote_uri: git_context.uri,
        git_provider: git_context.provider,
        branch: git_context.branch,
        commit_hash: git_context.id,
        envs: local_option.envs.unwrap_or_default(),
        ..Default::default()
    });

    debug!("Request for build: {:?}", &request);

    let response = client.build_start(request).await?;
    println!("RESPONSE = {:?}", response);

    Ok(())
}

use structopt::StructOpt;

use crate::GlobalOption;

use orbital_headers::build_meta::{build_service_client::BuildServiceClient, BuildTarget};
use orbital_headers::orbital_types::JobTrigger;

use chrono::NaiveDateTime;
use config_parser::yaml as parser;
use git_meta::git_info;
use orbital_database::postgres::schema::JobState;
use orbital_services::ORB_DEFAULT_URI;
use prettytable::{cell, format, row, Table};
use tonic::Request;

use anyhow::Result;
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

    /// Print full commit hash
    #[structopt(long, short)]
    wide: bool,
}

/// Generates gRPC `BuildStartRequest` object and connects to *currently hardcoded* gRPC server and sends a request to `BuildService` server.
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    local_option: SubcommandOption,
) -> Result<()> {
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
    //let _config = parser::load_orb_yaml(Path::new(&format!("{}/{}", &path.display(), "orb.yml")))?;
    let _config =
        parser::load_orb_yaml(Path::new(&format!("{}/{}", &path.display(), "orb.yml"))).unwrap();
    // Assuming Docker builder... (Stay focused!)
    // Get the docker container image

    // Org - default (Future: How can we cache this client-side?)
    // Validate that the org exists be

    let request = Request::new(BuildTarget {
        org: local_option.org.expect("Please provide an org name"),
        git_repo: git_context.git_url.name,
        remote_uri: git_context.git_url.href,
        //git_provider: git_context.git_url.host.unwrap(),
        branch: git_context.branch,
        commit_hash: git_context.commit_id,
        user_envs: local_option.envs.unwrap_or_default(),
        trigger: JobTrigger::Manual.into(),
        ..Default::default()
    });

    debug!("Request for build: {:?}", &request);

    let response = client.build_start(request).await?.into_inner();
    println!("RESPONSE = {:?}", response);

    // By default, format the response into a table
    let mut table = Table::new();
    table.set_format(*format::consts::FORMAT_NO_BORDER_LINE_SEPARATOR);

    // Print the header row
    table.set_titles(row![
        bc =>
        "Build #",
        "Org",
        "Repo",
        "Branch",
        "Commit",
        "User Envs",
        "Queue time",
        "Start time",
        "End time",
        "Build state",
    ]);

    let build_target = response.build.clone().expect("No build target in summary");

    let commit = match local_option.wide {
        true => build_target.commit_hash,
        false => build_target.commit_hash[..7].to_string(),
    };

    let queue_time = match &response.queue_time {
        Some(t) => format!(
            "{:?}",
            NaiveDateTime::from_timestamp(t.seconds, t.nanos as u32)
        ),
        None => format!("---"),
    };

    let start_time = match &response.start_time {
        Some(t) => format!(
            "{:?}",
            NaiveDateTime::from_timestamp(t.seconds, t.nanos as u32)
        ),
        None => format!("---"),
    };

    let end_time = match &response.end_time {
        Some(t) => format!(
            "{:?}",
            NaiveDateTime::from_timestamp(t.seconds, t.nanos as u32)
        ),
        None => format!("---"),
    };

    table.add_row(row![
        response.id,
        build_target.org,
        build_target.git_repo,
        build_target.branch,
        commit,
        build_target.user_envs,
        queue_time,
        start_time,
        end_time,
        JobState::from(response.build_state),
    ]);

    // Print the table to stdout
    table.printstd();

    Ok(())
}

use structopt::StructOpt;

use crate::GlobalOption;

use orbital_headers::build_meta::{
    build_service_client::BuildServiceClient, BuildMetadata, BuildTarget,
};
use orbital_headers::orbital_types::JobTrigger;

use chrono::NaiveDateTime;
use config_parser::yaml as parser;
use git_meta::git_info;
use orbital_database::postgres::schema::JobState;
use orbital_services::ORB_DEFAULT_URI;
use prettytable::{cell, format, row, Table};
use tonic::Request;

use std::io::Write;
use termcolor::{BufferWriter, ColorChoice};

use anyhow::Result;
use log::debug;
use std::path::{Path, PathBuf};
use std::str;

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

    /// Don't follow the build output
    #[structopt(long)]
    no_follow: bool,

    /// Path to local build config file to use instead of the checked in orb.yml
    #[structopt(long, parse(from_os_str))]
    config: Option<PathBuf>,
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

    let config = match &local_option.config {
        Some(path) => parser::load_orb_yaml(path).expect("Provided config failed validation"),
        None => parser::load_orb_yaml(Path::new(&format!("{}/{}", &path.display(), "orb.yml")))
            .expect("orb.yml failed validation"),
    };

    //let _config =
    //    parser::load_orb_yaml(Path::new(&format!("{}/{}", &path.display(), "orb.yml"))).unwrap();

    // Assuming Docker builder... (Stay focused!)
    // Get the docker container image

    // Org - default (Future: How can we cache this client-side?)
    // Validate that the org exists be

    let request = Request::new(BuildTarget {
        org: local_option.org.expect("Please provide an org name"),
        git_repo: git_context.git_url.name.clone(),
        remote_uri: git_context.git_url.trim_auth().to_string(),
        branch: git_context.branch,
        commit_hash: git_context.commit_id,
        user_envs: local_option.envs.unwrap_or_default(),
        trigger: JobTrigger::Manual.into(),
        config: {
            match local_option.config {
                Some(_path) => config.to_string(),
                None => "".to_string(),
            }
        },
        ..Default::default()
    });

    debug!("Request for build: {:?}", &request);

    //let response = client.build_start(request).await?.into_inner();
    //println!("RESPONSE = {:?}", response);

    let mut stream = client.build_start(request).await?.into_inner();

    let mut build_metadata = BuildMetadata {
        ..Default::default()
    };

    // TODO: Convert the proto types to the Diesel types
    //while let Some(response) = stream.next().await {
    while let Some(response) = stream.message().await? {
        let bufwtr = BufferWriter::stdout(ColorChoice::Auto);
        let mut buffer = bufwtr.buffer();

        build_metadata = response.build_metadata.clone().unwrap_or_default();

        match response.build_output.clone().pop() {
            Some(build_output) => {
                //writeln!(&mut buffer, "{:?}", response.clone())?;
                writeln!(
                    &mut buffer,
                    "[{}] {}",
                    build_output.stage.to_string(),
                    str::from_utf8(&build_output.output)?
                )?;
                bufwtr.print(&buffer)?;
            }
            None => (),
        };
    }

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

    let build_target = build_metadata
        .build
        .clone()
        .expect("No build target in summary");

    let commit = match local_option.wide {
        true => build_target.commit_hash,
        false => build_target.commit_hash[..7].to_string(),
    };

    let queue_time = match &build_metadata.queue_time {
        Some(t) => format!(
            "{:?}",
            NaiveDateTime::from_timestamp(t.seconds, t.nanos as u32)
        ),
        None => format!("---"),
    };

    let start_time = match &build_metadata.start_time {
        Some(t) => format!(
            "{:?}",
            NaiveDateTime::from_timestamp(t.seconds, t.nanos as u32)
        ),
        None => format!("---"),
    };

    let end_time = match &build_metadata.end_time {
        Some(t) => format!(
            "{:?}",
            NaiveDateTime::from_timestamp(t.seconds, t.nanos as u32)
        ),
        None => format!("---"),
    };

    table.add_row(row![
        build_metadata.id,
        build_target.org,
        build_target.git_repo,
        build_target.branch,
        commit,
        build_target.user_envs,
        queue_time,
        start_time,
        end_time,
        JobState::from(build_metadata.build_state),
    ]);

    // Print the table to stdout
    table.printstd();

    Ok(())
}

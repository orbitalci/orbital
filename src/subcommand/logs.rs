use structopt::StructOpt;

use crate::subcommand::GlobalOption;

use crate::orbital_headers::build_meta::{build_service_client::BuildServiceClient, BuildTarget};

use crate::orbital_services::ORB_DEFAULT_URI;
use color_eyre::eyre::Result;
use git_meta::GitRepo;
use std::io::Write;
use std::path::PathBuf;
use termcolor::{BufferWriter, ColorChoice};
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

    /// Build ID
    #[structopt(long)]
    id: Option<i32>,
}

/// Generates request for logs
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    local_option: SubcommandOption,
) -> Result<()> {
    let mut client = BuildServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    // Path
    let path = local_option.path;

    let git_context = GitRepo::open(path, local_option.branch, local_option.hash)
        .expect("Unable to open GitRepo");
    // Need to figure out how to handle the streaming response

    let request = Request::new(BuildTarget {
        org: local_option.org.expect("Please provide an org name"),
        git_repo: git_context.url.name.clone(),
        remote_uri: git_context.url.trim_auth().to_string(),
        branch: git_context.branch.unwrap_or_default(),
        commit_hash: git_context.head.unwrap().id,
        user_envs: local_option.envs.unwrap_or_default(),
        id: local_option.id.unwrap_or(0),
        ..Default::default()
    });

    let mut stream = client.build_logs(request).await?.into_inner();

    while let Some(response) = stream.message().await? {
        let bufwtr = BufferWriter::stdout(ColorChoice::Auto);
        let mut buffer = bufwtr.buffer();

        for records in response.records {
            for logs in records.build_output {
                writeln!(
                    &mut buffer,
                    "{}",
                    &String::from_utf8(logs.output.clone()).unwrap()
                )?;
                bufwtr.print(&buffer)?;
            }
        }
    }

    Ok(())
}

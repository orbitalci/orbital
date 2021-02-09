use structopt::StructOpt;

use crate::{repo::SubcommandOption, GlobalOption};

use orbital_headers::code::{code_service_client::CodeServiceClient, GitRepoGetRequest};
use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

use git_meta::GitRepo;
use log::debug;
use std::path::PathBuf;

use anyhow::Result;
use orbital_database::postgres::repo::Repo;
use prettytable::{cell, format, row, Table};

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    /// Repo path
    #[structopt(parse(from_os_str), env = "PWD")]
    path: PathBuf,

    /// Repo name
    #[structopt(long)]
    name: Option<String>,

    /// Name of Orbital org
    #[structopt(long, env = "ORB_DEFAULT_ORG")]
    org: Option<String>,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<()> {
    // Try to parse path for some git info

    let repo_info = GitRepo::open(action_option.path, None, None).expect("Unable to open GitRepo");

    let request = match action_option.name {
        Some(name) => Request::new(GitRepoGetRequest {
            org: action_option
                .org
                .clone()
                .expect("Please provide an org with request"),
            name: name,
            ..Default::default()
        }),
        None => Request::new(GitRepoGetRequest {
            org: action_option
                .org
                .clone()
                .expect("Please provide an org with request"),
            git_provider: repo_info.url.host.clone().unwrap(),
            name: repo_info.url.name.clone(),
            user: repo_info.url.user.clone().unwrap(),
            uri: repo_info.url.trim_auth().to_string(),
        }),
    };

    debug!("Request for git repo get: {:?}", &request);

    let mut client = CodeServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let response = client.git_repo_get(request).await;

    match response {
        Err(_e) => {
            eprintln!("Repo not found");
            Ok(())
        }
        Ok(o) => {
            let repo_proto = o.into_inner();

            debug!("RESPONSE = {:?}", &repo_proto);

            // By default, format the response into a table
            let mut table = Table::new();
            table.set_format(*format::consts::FORMAT_NO_BORDER_LINE_SEPARATOR);

            // Print the header row
            table.set_titles(row![
                bc =>
                "Org Name",
                "Repo Name",
                "Uri",
                "Secret Name",
                "Build Enabled",
                "Notify Enabled",
                "Next build index"
            ]);

            let repo = Repo::from(repo_proto.clone());

            table.add_row(row![
                action_option.org.unwrap(),
                repo.name,
                repo.uri,
                repo_proto.auth_data,
                &format!("{:?}", repo.build_active_state),
                &format!("{:?}", repo.notify_active_state),
                repo.next_build_index
            ]);

            // Print the table to stdout
            table.printstd();
            Ok(())
        }
    }
}

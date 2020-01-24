use structopt::StructOpt;

use crate::{repo::SubcommandOption, GlobalOption, SubcommandError};

use orbital_headers::code::{code_service_client::CodeServiceClient, GitRepoListRequest};
use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

use log::debug;

use orbital_database::postgres::repo::Repo;
use prettytable::{cell, format, row, Table};

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    /// Name of Orbital org
    #[structopt(long, env = "ORB_DEFAULT_ORG")]
    org: Option<String>,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<(), SubcommandError> {
    let mut client = CodeServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let request = Request::new(GitRepoListRequest {
        org: action_option.org.unwrap_or_default(),
        ..Default::default()
    });

    debug!("Request for git repo list: {:?}", &request);

    let response = client.git_repo_list(request).await?.into_inner();

    // By default, format the response into a table
    let mut table = Table::new();
    table.set_format(*format::consts::FORMAT_NO_BORDER_LINE_SEPARATOR);

    // Print the header row
    table.set_titles(row![
        "Org Name",
        "Repo Name",
        "Uri",
        "Secret Type",
        "Build Enabled",
        "Notify Enabled",
        "Next build index"
    ]);

    match response.git_repos.len() {
        0 => {
            println!("No repos found");
        }
        _ => {
            for repo_proto in &response.git_repos {
                let repo = Repo::from(repo_proto.clone());

                table.add_row(row![
                    repo.org_id,
                    repo.name,
                    repo.uri,
                    repo.secret_id.unwrap_or_default(),
                    &format!("{:?}", repo.build_active_state),
                    &format!("{:?}", repo.notify_active_state),
                    repo.next_build_index
                ]);

                debug!("RESPONSE = {:?}", &response);
            }
        }
    }

    // Print the table to stdout
    table.printstd();

    Ok(())
}

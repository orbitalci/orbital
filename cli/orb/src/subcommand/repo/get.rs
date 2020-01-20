use structopt::StructOpt;

use crate::{repo::SubcommandOption, GlobalOption, SubcommandError};

use orbital_headers::code::{code_service_client::CodeServiceClient, GitRepoGetRequest};
use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

use git_meta::git_info;
use log::debug;
use std::path::PathBuf;

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
) -> Result<(), SubcommandError> {
    // Try to parse path for some git info

    let repo_info =
        match git_info::get_git_info_from_path(&action_option.path.as_path(), &None, &None) {
            Ok(info) => info,
            Err(_e) => panic!("Unable to parse path for git repo info"),
        };

    let request = match action_option.name {
        Some(name) => Request::new(GitRepoGetRequest {
            org: action_option.org.unwrap_or_default(),
            name: name,
            ..Default::default()
        }),
        None => Request::new(GitRepoGetRequest {
            org: action_option.org.unwrap_or_default(),
            ..Default::default()
        }),
    };

    debug!("Request for git repo get: {:?}", &request);

    let mut client = CodeServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    let response = client.git_repo_get(request).await?;
    println!("RESPONSE = {:?}", response);
    Ok(())
}

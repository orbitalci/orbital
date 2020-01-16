use structopt::StructOpt;

use crate::{repo::SubcommandOption, GlobalOption, SubcommandError};

use orbital_headers::code::{code_service_client::CodeServiceClient, GitRepoAddRequest};
use orbital_headers::orbital_types::SecretType;
use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

use git_meta::git_info;
use git_meta::GitSshRemote;
use log::debug;
use std::fs::File;
use std::io::prelude::*;
use std::path::PathBuf;

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    /// Repo path
    #[structopt(parse(from_os_str), env = "PWD")]
    path: PathBuf,

    /// Name of Orbital org
    #[structopt(long, env = "ORB_DEFAULT_ORG")]
    org: Option<String>,

    /// Use flag if repo is public
    #[structopt(long)]
    public: bool,

    // TODO: We're only supporting ssh key auth from the client at the moment.
    /// Path to private key
    #[structopt(long, parse(from_os_str), required_if("public", "false"))]
    private_key: PathBuf,

    /// Username for private repo
    #[structopt(long, short = "u")]
    username: Option<String>,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<(), SubcommandError> {
    let mut client = CodeServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;

    // Read git repo info from local filesystem.
    // TODO: support adding a repo from a uri
    let repo_info =
        match git_info::get_git_info_from_path(&action_option.path.as_path(), &None, &None) {
            Ok(info) => info,
            Err(_e) => panic!("Unable to parse path for git repo info"),
        };

    // TODO: Need to update the git repo parser to split out a username
    let request = match &action_option.public {
        true => Request::new(GitRepoAddRequest {
            org: action_option.org.unwrap_or_default(),
            secret_type: SecretType::Unspecified.into(),
            git_provider: repo_info.provider,
            name: format!("{}/{}", repo_info.account, repo_info.repo),
            uri: repo_info.uri.clone(),
            user: git_info::git_remote_url_parse(&repo_info.uri.clone()).user,
            ..Default::default()
        }),
        false => {
            // Read in private key into memory
            let mut file = File::open(
                &action_option
                    .private_key
                    .to_str()
                    .expect("No secret filepath given"),
            )?;
            let mut contents = String::new();
            file.read_to_string(&mut contents)?;

            Request::new(GitRepoAddRequest {
                org: action_option.org.unwrap_or_default(),
                secret_type: SecretType::SshKey.into(),
                auth_data: contents,
                git_provider: repo_info.provider,
                name: format!("{}/{}", repo_info.account, repo_info.repo),
                uri: repo_info.uri.clone(),
                user: git_info::git_remote_url_parse(&repo_info.uri.clone()).user,
            })
        }
    };

    debug!("Request for git repo add: {:?}", &request);

    let response = client.git_repo_add(request).await?;
    println!("RESPONSE = {:?}", response);
    Ok(())
}

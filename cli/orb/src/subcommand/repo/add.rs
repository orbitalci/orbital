use structopt::StructOpt;

use crate::{repo::SubcommandOption, GlobalOption};

use orbital_headers::code::{code_service_client::CodeServiceClient, GitRepoAddRequest};
use orbital_headers::orbital_types::SecretType;
use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

use git_meta::{git_info, GitCredentials};
use git_url_parse::Scheme;
use log::{debug, info};
use std::fs::File;
use std::io::prelude::*;
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

    /// Name of Orbital org
    #[structopt(long, env = "ORB_DEFAULT_ORG")]
    org: Option<String>,

    /// Path to private key
    #[structopt(long, parse(from_os_str), conflicts_with("password"))]
    private_key: Option<PathBuf>,

    /// Password for private repo. Mutually exclusive w/ private key
    #[structopt(long, conflicts_with("private-key"))]
    password: Option<String>,

    /// Username for private repo
    #[structopt(long, short = "u")]
    username: Option<String>,

    /// Skip checking branch clone before adding
    #[structopt(long)]
    skip_check: bool,

    /// Check clone with provided branch instead of master
    #[structopt(long)]
    alt_branch: Option<String>,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<()> {
    // Read git repo info from local filesystem.
    // TODO: support adding a repo from a uri
    let repo_info =
        match git_info::get_git_info_from_path(&action_option.path.as_path(), &None, &None) {
            Ok(info) => info,
            Err(_e) => panic!("Unable to parse path for git repo info"),
        };

    info!("Adding repo: {:?}", &repo_info);

    let (repo_secret_type, repo_user) = match repo_info.git_url.scheme {
        Scheme::Ssh => (SecretType::SshKey, repo_info.git_url.user.clone().unwrap()),
        Scheme::Https => match &repo_info.git_url.token {
            Some(_) => (
                SecretType::BasicAuth,
                if let Some(user) = action_option.username {
                    user
                } else {
                    repo_info.git_url.user.clone().unwrap()
                },
            ),
            None => (
                SecretType::Unspecified,
                action_option.username.unwrap_or_default(),
            ),
        },
        _ => (
            SecretType::Unspecified,
            action_option.username.unwrap_or_default(),
        ),
    };

    let mut auth_content = String::new();

    // This is only for getting remote branch info before sending the add repo request
    let mut git_creds = GitCredentials::Public;

    if repo_secret_type != SecretType::Unspecified {
        // If private key, load up contents with key
        match action_option.private_key {
            Some(p) => {
                info!("Repo auth with private key");

                // Read in private key into memory
                let mut file = File::open(p.to_str().expect("No secret filepath given"))?;
                file.read_to_string(&mut auth_content)?;

                git_creds = GitCredentials::SshKey {
                    username: repo_user.clone(),
                    public_key: None,
                    private_key: auth_content.clone(),
                    passphrase: None,
                };
            }
            None => info!("Not using private key auth"),
        };

        // If password, load up contents with password
        match action_option.password {
            Some(p) => {
                info!("Repo auth with basic auth");

                auth_content = p;

                git_creds = GitCredentials::BasicAuth {
                    username: repo_user.clone(),
                    password: auth_content.clone(),
                }
            }
            None => info!("Not using basic auth"),
        };
    }

    // Retrieve the latest remote branch refs
    let remote_branch_refs =
        git_info::list_remote_branch_head_refs(&action_option.path.as_path(), git_creds)?;

    println!("{:?}", remote_branch_refs);

    let request = Request::new(GitRepoAddRequest {
        org: action_option
            .org
            .clone()
            .expect("Please provide an org with request"),
        secret_type: repo_secret_type.into(),
        auth_data: auth_content,
        git_provider: repo_info.git_url.host.clone().unwrap(),
        name: repo_info.git_url.name.clone(),
        uri: repo_info.git_url.trim_auth().to_string(),
        user: repo_user,
        alt_check_branch: action_option.alt_branch.unwrap_or_default(),
        skip_check: action_option.skip_check,
        ..Default::default()
    });

    debug!("Request for git repo add: {:?}", &request);

    let mut client = CodeServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;
    let response = client.git_repo_add(request).await;

    match response {
        Err(_e) => {
            eprintln!("Error adding Repo");
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
